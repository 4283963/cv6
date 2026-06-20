const DEFAULT_PORT = 37245

async function getPort(): Promise<number> {
  if (window.electronAPI?.getServerPort) {
    try {
      return await window.electronAPI.getServerPort()
    } catch {
      return DEFAULT_PORT
    }
  }
  return DEFAULT_PORT
}

async function baseUrl(): Promise<string> {
  const port = await getPort()
  return `http://127.0.0.1:${port}`
}

export interface NoteNode {
  id: string
  title: string
  path: string
  modTime: number
  links: string[]
  linkFrom?: string[]
}

export interface NoteLink {
  source: string
  target: string
}

export interface GraphData {
  nodes: NoteNode[]
  links: NoteLink[]
}

export interface ScanResponse {
  success: boolean
  message?: string
  data?: GraphData
}

export interface ReadFileResponse {
  success: boolean
  message?: string
  content?: string
  title?: string
}

export async function scanFolder(folder: string): Promise<ScanResponse> {
  const base = await baseUrl()
  const res = await fetch(`${base}/api/scan`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ folder }),
  })
  return res.json()
}

export async function readNoteFile(path: string): Promise<ReadFileResponse> {
  const base = await baseUrl()
  const res = await fetch(`${base}/api/read`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ path }),
  })
  return res.json()
}

export async function healthCheck(): Promise<boolean> {
  try {
    const base = await baseUrl()
    const res = await fetch(`${base}/api/health`)
    return res.ok
  } catch {
    return false
  }
}
