package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type NoteFile struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Path     string   `json:"path"`
	ModTime  int64    `json:"modTime"`
	Links    []string `json:"links"`
	LinkFrom []string `json:"linkFrom,omitempty"`
}

type GraphData struct {
	Nodes []NoteFile `json:"nodes"`
	Links []Link     `json:"links"`
}

type Link struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type ScanRequest struct {
	Folder string `json:"folder"`
}

type ScanResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message,omitempty"`
	Data    GraphData `json:"data,omitempty"`
}

type ReadFileRequest struct {
	Path string `json:"path"`
}

type ReadFileResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Content string `json:"content,omitempty"`
	Title   string `json:"title,omitempty"`
}

var (
	linkRegex  = regexp.MustCompile(`\[\[([^\[\]\n]+?)\]\]`)
	cacheMu    sync.RWMutex
	lastScan   string
	cachedData GraphData
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/scan", handleScan)
	mux.HandleFunc("/api/read", handleReadFile)
	mux.HandleFunc("/api/health", handleHealth)
	mux.HandleFunc("/", handleCORS)

	port := os.Getenv("PORT")
	if port == "" {
		port = "37245"
	}

	addr := fmt.Sprintf("127.0.0.1:%s", port)
	log.Printf("MD Note Graph server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func handleCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func handleScan(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ScanResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Folder == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ScanResponse{
			Success: false,
			Message: "Folder path is required",
		})
		return
	}

	absFolder, err := filepath.Abs(req.Folder)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ScanResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid folder path: %v", err),
		})
		return
	}

	cacheMu.RLock()
	if lastScan == absFolder && len(cachedData.Nodes) > 0 {
		cacheMu.RUnlock()
		json.NewEncoder(w).Encode(ScanResponse{
			Success: true,
			Data:    cachedData,
		})
		return
	}
	cacheMu.RUnlock()

	data, err := scanFolder(absFolder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ScanResponse{
			Success: false,
			Message: fmt.Sprintf("Scan failed: %v", err),
		})
		return
	}

	cacheMu.Lock()
	lastScan = absFolder
	cachedData = data
	cacheMu.Unlock()

	json.NewEncoder(w).Encode(ScanResponse{
		Success: true,
		Data:    data,
	})
}

func handleReadFile(w http.ResponseWriter, r *http.Request) {
	setCORSHeaders(w)
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req ReadFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReadFileResponse{
			Success: false,
			Message: "Invalid request body",
		})
		return
	}

	if req.Path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ReadFileResponse{
			Success: false,
			Message: "File path is required",
		})
		return
	}

	content, err := os.ReadFile(req.Path)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ReadFileResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to read file: %v", err),
		})
		return
	}

	title := filepath.Base(req.Path)
	title = strings.TrimSuffix(title, filepath.Ext(title))

	json.NewEncoder(w).Encode(ReadFileResponse{
		Success: true,
		Content: string(content),
		Title:   title,
	})
}

func scanFolder(folder string) (GraphData, error) {
	var data GraphData
	noteMap := make(map[string]*NoteFile)
	titleToID := make(map[string]string)

	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(folder, path)
		id := relPath
		title := strings.TrimSuffix(filepath.Base(path), ".md")

		info, err := d.Info()
		var modTime int64
		if err == nil {
			modTime = info.ModTime().Unix()
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Warning: cannot read %s: %v", path, err)
			return nil
		}

		rawLinks := linkRegex.FindAllStringSubmatch(string(content), -1)
		links := make([]string, 0, len(rawLinks))
		seen := make(map[string]bool)
		for _, match := range rawLinks {
			if len(match) > 1 {
				linkText := strings.TrimSpace(match[1])
				pipeIdx := strings.Index(linkText, "|")
				if pipeIdx > 0 {
					linkText = strings.TrimSpace(linkText[:pipeIdx])
				}
				hashIdx := strings.Index(linkText, "#")
				if hashIdx > 0 {
					linkText = strings.TrimSpace(linkText[:hashIdx])
				}
				if linkText != "" && !seen[linkText] {
					seen[linkText] = true
					links = append(links, linkText)
				}
			}
		}

		note := &NoteFile{
			ID:      id,
			Title:   title,
			Path:    path,
			ModTime: modTime,
			Links:   links,
		}
		noteMap[id] = note
		titleToID[title] = id
		titleToID[strings.ToLower(title)] = id
		titleToID[relPath] = id
		titleToID[strings.TrimSuffix(relPath, ".md")] = id

		return nil
	})

	if err != nil {
		return data, err
	}

	linkFromSetMap := make(map[string]map[string]bool)
	for id := range noteMap {
		linkFromSetMap[id] = make(map[string]bool)
	}

	for _, note := range noteMap {
		resolvedLinks := make([]string, 0, len(note.Links))
		resolvedSet := make(map[string]bool)
		for _, link := range note.Links {
			if targetID, ok := resolveLink(link, titleToID, noteMap); ok {
				if resolvedSet[targetID] {
					continue
				}
				resolvedSet[targetID] = true
				resolvedLinks = append(resolvedLinks, targetID)
				if target, exists := noteMap[targetID]; exists {
					targetLinkFromSet, _ := linkFromSetMap[targetID]
					if !targetLinkFromSet[note.ID] {
						targetLinkFromSet[note.ID] = true
						target.LinkFrom = append(target.LinkFrom, note.ID)
					}
				}
			}
		}
		note.Links = resolvedLinks
	}

	data.Nodes = make([]NoteFile, 0, len(noteMap))
	for _, note := range noteMap {
		data.Nodes = append(data.Nodes, *note)
	}

	linkSet := make(map[string]bool)
	data.Links = make([]Link, 0)
	for _, note := range noteMap {
		for _, target := range note.Links {
			key := note.ID + "->" + target
			if !linkSet[key] {
				linkSet[key] = true
				data.Links = append(data.Links, Link{
					Source: note.ID,
					Target: target,
				})
			}
		}
	}

	return data, nil
}

func resolveLink(link string, titleToID map[string]string, noteMap map[string]*NoteFile) (string, bool) {
	cleanLink := strings.TrimSpace(link)
	cleanLink = strings.TrimSuffix(cleanLink, ".md")

	if id, ok := titleToID[cleanLink]; ok {
		return id, true
	}
	if id, ok := titleToID[strings.ToLower(cleanLink)]; ok {
		return id, true
	}

	for id, note := range noteMap {
		if strings.EqualFold(note.Title, cleanLink) {
			return id, true
		}
		noteBase := strings.TrimSuffix(filepath.Base(note.Path), ".md")
		if strings.EqualFold(noteBase, cleanLink) {
			return id, true
		}
	}

	return "", false
}
