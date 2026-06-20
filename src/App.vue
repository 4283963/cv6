<template>
  <div class="app-shell">
    <header class="top-bar">
      <div class="brand">
        <div class="logo-dot"></div>
        <span class="brand-name">MD Note Graph</span>
      </div>

      <div class="folder-section">
        <button class="btn btn-primary" @click="handleSelectFolder">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          </svg>
          选择文件夹
        </button>

        <div v-if="currentFolder" class="folder-path">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"></path>
          </svg>
          <span class="path-text" :title="currentFolder">{{ currentFolder }}</span>
        </div>
      </div>

      <div class="controls">
        <button class="btn btn-ghost" :disabled="!graphData" @click="handleRefresh">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="23 4 23 10 17 10"></polyline>
            <polyline points="1 20 1 14 7 14"></polyline>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path>
          </svg>
          刷新
        </button>
        <button class="btn btn-ghost" :disabled="!graphData" @click="handleResetView">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"></path>
            <path d="M3 3v5h5"></path>
          </svg>
          重置视图
        </button>
      </div>
    </header>

    <div class="main-area">
      <aside class="sidebar">
        <div class="sidebar-header">
          <h3>笔记列表</h3>
          <div v-if="graphData" class="node-count">{{ graphData.nodes.length }} 个笔记</div>
        </div>

        <div class="search-box">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
          <input v-model="searchQuery" type="text" placeholder="搜索笔记..." />
        </div>

        <div class="note-list" v-if="filteredNodes.length > 0">
          <div
            v-for="node in filteredNodes"
            :key="node.id"
            class="note-item"
            :class="{ active: selectedNode?.id === node.id }"
            @click="handleSelectNode(node)"
          >
            <div class="note-title">{{ node.title }}</div>
            <div class="note-meta">
              <span class="badge badge-out">出 {{ node.links?.length || 0 }}</span>
              <span class="badge badge-in">入 {{ node.linkFrom?.length || 0 }}</span>
            </div>
          </div>
        </div>
        <div v-else class="empty-sidebar">
          <template v-if="!currentFolder">
            <div class="empty-icon">📁</div>
            <p>请选择一个笔记文件夹</p>
          </template>
          <template v-else-if="loading">
            <div class="empty-icon">⏳</div>
            <p>正在扫描中...</p>
          </template>
          <template v-else-if="searchQuery">
            <div class="empty-icon">🔍</div>
            <p>没有匹配的笔记</p>
          </template>
          <template v-else>
            <div class="empty-icon">📝</div>
            <p>未找到 .md 文件</p>
          </template>
        </div>
      </aside>

      <section class="graph-area">
        <div v-if="loading" class="loading-overlay">
          <div class="spinner"></div>
          <p>正在扫描笔记并构建关系图...</p>
        </div>

        <div v-else-if="error" class="error-overlay">
          <div class="error-icon">⚠️</div>
          <p class="error-msg">{{ error }}</p>
        </div>

        <div v-else-if="!graphData || graphData.nodes.length === 0" class="empty-overlay">
          <div class="empty-illustration">
            <svg width="120" height="120" viewBox="0 0 120 120" fill="none">
              <circle cx="60" cy="60" r="50" stroke="#2d3340" stroke-width="2" stroke-dasharray="4 6" />
              <circle cx="40" cy="45" r="8" fill="#3d4658" />
              <circle cx="80" cy="45" r="6" fill="#3d4658" />
              <circle cx="60" cy="80" r="10" fill="#3d4658" />
              <circle cx="35" cy="75" r="5" fill="#3d4658" />
              <line x1="40" y1="45" x2="60" y2="80" stroke="#3d4658" stroke-width="1.5" />
              <line x1="80" y1="45" x2="60" y2="80" stroke="#3d4658" stroke-width="1.5" />
              <line x1="40" y1="45" x2="80" y2="45" stroke="#3d4658" stroke-width="1.5" stroke-dasharray="3 3" />
              <line x1="35" y1="75" x2="60" y2="80" stroke="#3d4658" stroke-width="1.5" />
            </svg>
          </div>
          <h2>Markdown 笔记关系图</h2>
          <p>点击「选择文件夹」开始扫描你的笔记</p>
          <p class="hint">工具会解析文件中的 [[链接]] 语法，自动构建笔记间的引用网络</p>
        </div>

        <NoteGraph
          v-else
          ref="graphRef"
          :data="graphData"
          :selected-id="selectedNode?.id || null"
          @select="handleSelectNode"
          @open="handleOpenNode"
        />
      </section>

      <aside v-if="selectedNode" class="detail-panel">
        <div class="detail-header">
          <h3 class="detail-title">{{ selectedNode.title }}</h3>
          <button class="close-btn" @click="selectedNode = null">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>

        <div class="detail-stats">
          <div class="stat">
            <div class="stat-value">{{ selectedNode.links?.length || 0 }}</div>
            <div class="stat-label">出链</div>
          </div>
          <div class="stat">
            <div class="stat-value">{{ selectedNode.linkFrom?.length || 0 }}</div>
            <div class="stat-label">入链</div>
          </div>
        </div>

        <div class="detail-section">
          <h4>文件路径</h4>
          <div class="detail-path">{{ selectedNode.path }}</div>
        </div>

        <div v-if="noteContent !== null" class="detail-section">
          <h4>内容预览</h4>
          <div class="detail-content">
            <pre>{{ noteContent }}</pre>
          </div>
        </div>
        <div v-else class="detail-section">
          <h4>内容预览</h4>
          <button class="btn btn-small btn-primary" @click="handleLoadContent">加载内容</button>
        </div>

        <div v-if="linkedNodes.length > 0" class="detail-section">
          <h4>相关笔记 ({{ linkedNodes.length }})</h4>
          <div class="linked-list">
            <div
              v-for="n in linkedNodes"
              :key="n.id"
              class="linked-item"
              @click="handleSelectNode(n)"
            >
              {{ n.title }}
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import NoteGraph from './components/NoteGraph.vue'
import { scanFolder, readNoteFile, type GraphData, type NoteNode } from './api'

const graphRef = ref<InstanceType<typeof NoteGraph> | null>(null)
const currentFolder = ref<string | null>(null)
const graphData = ref<GraphData | null>(null)
const selectedNode = ref<NoteNode | null>(null)
const searchQuery = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
const noteContent = ref<string | null>(null)

const filteredNodes = computed(() => {
  if (!graphData.value) return []
  if (!searchQuery.value) {
    return [...graphData.value.nodes].sort((a, b) => {
      const degA = (a.links?.length || 0) + (a.linkFrom?.length || 0)
      const degB = (b.links?.length || 0) + (b.linkFrom?.length || 0)
      return degB - degA
    })
  }
  const q = searchQuery.value.toLowerCase()
  return graphData.value.nodes.filter(
    (n) => n.title.toLowerCase().includes(q) || n.id.toLowerCase().includes(q)
  )
})

const linkedNodes = computed(() => {
  if (!selectedNode.value || !graphData.value) return []
  const ids = new Set<string>([...(selectedNode.value.links || []), ...(selectedNode.value.linkFrom || [])])
  const nodeMap = new Map(graphData.value.nodes.map((n) => [n.id, n]))
  const result: NoteNode[] = []
  ids.forEach((id) => {
    const n = nodeMap.get(id)
    if (n) result.push(n)
  })
  return result.sort((a, b) => a.title.localeCompare(b.title))
})

async function handleSelectFolder() {
  if (!window.electronAPI?.selectFolder) {
    const fallback = prompt('请输入文件夹绝对路径（开发模式）:')
    if (fallback) {
      currentFolder.value = fallback
      await doScan()
    }
    return
  }
  const folder = await window.electronAPI.selectFolder()
  if (folder) {
    currentFolder.value = folder
    selectedNode.value = null
    noteContent.value = null
    await doScan()
  }
}

async function doScan() {
  if (!currentFolder.value) return
  loading.value = true
  error.value = null
  try {
    const res = await scanFolder(currentFolder.value)
    if (res.success && res.data) {
      graphData.value = res.data
    } else {
      error.value = res.message || '扫描失败'
      graphData.value = null
    }
  } catch (e: any) {
    error.value = `无法连接后端服务: ${e.message || e}`
    graphData.value = null
  } finally {
    loading.value = false
  }
}

async function handleRefresh() {
  await doScan()
}

function handleResetView() {
  graphRef.value?.resetView()
}

function handleSelectNode(node: NoteNode) {
  if (!node || !node.id) {
    selectedNode.value = null
    return
  }
  selectedNode.value = node
  noteContent.value = null
}

function handleOpenNode(node: NoteNode) {
  selectedNode.value = node
  handleLoadContent()
}

async function handleLoadContent() {
  if (!selectedNode.value) return
  try {
    const res = await readNoteFile(selectedNode.value.path)
    if (res.success && res.content !== undefined) {
      noteContent.value = res.content
    }
  } catch {
    noteContent.value = '加载失败'
  }
}

watch(selectedNode, () => {
  noteContent.value = null
})
</script>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #0f1115;
}

.top-bar {
  display: flex;
  align-items: center;
  padding: 12px 20px;
  background: #171a22;
  border-bottom: 1px solid #242936;
  gap: 20px;
  flex-shrink: 0;
}

.brand {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-shrink: 0;
}

.logo-dot {
  width: 14px;
  height: 14px;
  background: linear-gradient(135deg, #7aa2f7, #bb9af7);
  border-radius: 4px;
}

.brand-name {
  font-weight: 600;
  font-size: 15px;
  color: #e8e8e8;
  letter-spacing: 0.3px;
}

.folder-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.folder-path {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: #1d212c;
  border: 1px solid #2a2f3d;
  border-radius: 6px;
  color: #8892a6;
  font-size: 12px;
  max-width: 420px;
  flex-shrink: 1;
  min-width: 0;
}

.path-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  border: 1px solid transparent;
  transition: all 0.15s ease;
  background: transparent;
  color: inherit;
  font-family: inherit;
}

.btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.btn-primary {
  background: linear-gradient(135deg, #5a87e8, #7aa2f7);
  color: #ffffff;
  border-color: #5a87e8;
}

.btn-primary:hover:not(:disabled) {
  background: linear-gradient(135deg, #6a93ec, #89aef9);
}

.btn-ghost {
  color: #a8b0c0;
  border-color: #2d3340;
  background: transparent;
}

.btn-ghost:hover:not(:disabled) {
  background: #1d212c;
  color: #e8e8e8;
}

.btn-small {
  padding: 5px 10px;
  font-size: 12px;
}

.main-area {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.sidebar {
  width: 280px;
  background: #13161e;
  border-right: 1px solid #242936;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 16px 10px;
}

.sidebar-header h3 {
  font-size: 13px;
  font-weight: 600;
  color: #c8cfdb;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

.node-count {
  font-size: 11px;
  color: #656d80;
  background: #1d212c;
  padding: 2px 8px;
  border-radius: 10px;
}

.search-box {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px 12px;
  border-bottom: 1px solid #1f242f;
}

.search-box svg {
  color: #656d80;
  flex-shrink: 0;
}

.search-box input {
  flex: 1;
  background: #1a1e28;
  border: 1px solid #2a2f3d;
  border-radius: 6px;
  padding: 7px 10px;
  color: #e8e8e8;
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s;
  font-family: inherit;
}

.search-box input:focus {
  border-color: #5a87e8;
}

.search-box input::placeholder {
  color: #50586a;
}

.note-list {
  flex: 1;
  overflow-y: auto;
  padding: 6px 0;
}

.note-item {
  padding: 10px 16px;
  cursor: pointer;
  transition: background 0.1s;
  border-left: 2px solid transparent;
}

.note-item:hover {
  background: #1a1e28;
}

.note-item.active {
  background: #1e2638;
  border-left-color: #7aa2f7;
}

.note-title {
  font-size: 13px;
  color: #e8e8e8;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.note-meta {
  display: flex;
  gap: 6px;
}

.badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 500;
}

.badge-out {
  background: rgba(122, 162, 247, 0.15);
  color: #7aa2f7;
}

.badge-in {
  background: rgba(187, 154, 247, 0.15);
  color: #bb9af7;
}

.empty-sidebar {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
  text-align: center;
  color: #656d80;
  gap: 8px;
}

.empty-icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.empty-sidebar p {
  font-size: 13px;
}

.graph-area {
  flex: 1;
  position: relative;
  min-width: 0;
  min-height: 0;
}

.loading-overlay,
.error-overlay,
.empty-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 14px;
  padding: 40px;
  text-align: center;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #2d3340;
  border-top-color: #7aa2f7;
  border-radius: 50%;
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-overlay p {
  color: #8892a6;
  font-size: 14px;
}

.error-icon {
  font-size: 40px;
}

.error-msg {
  color: #f7768e;
  font-size: 14px;
  max-width: 480px;
}

.empty-illustration {
  margin-bottom: 12px;
  opacity: 0.7;
}

.empty-overlay h2 {
  color: #e8e8e8;
  font-size: 20px;
  font-weight: 600;
}

.empty-overlay p {
  color: #8892a6;
  font-size: 14px;
}

.empty-overlay .hint {
  color: #50586a;
  font-size: 12px;
}

.detail-panel {
  width: 340px;
  background: #13161e;
  border-left: 1px solid #242936;
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  overflow: hidden;
}

.detail-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 16px 16px 12px;
  gap: 12px;
  border-bottom: 1px solid #1f242f;
}

.detail-title {
  font-size: 16px;
  font-weight: 600;
  color: #e8e8e8;
  word-break: break-word;
  line-height: 1.4;
}

.close-btn {
  background: transparent;
  border: none;
  color: #656d80;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.15s;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.close-btn:hover {
  background: #1d212c;
  color: #e8e8e8;
}

.detail-stats {
  display: flex;
  gap: 1px;
  background: #1f242f;
  padding: 1px;
  margin: 12px 16px;
  border-radius: 8px;
}

.stat {
  flex: 1;
  text-align: center;
  padding: 10px 8px;
  background: #171a22;
  border-radius: 7px;
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #7aa2f7;
}

.stat-label {
  font-size: 11px;
  color: #656d80;
  margin-top: 2px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.detail-section {
  padding: 12px 16px;
  border-bottom: 1px solid #1a1e28;
}

.detail-section h4 {
  font-size: 11px;
  font-weight: 600;
  color: #656d80;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  margin-bottom: 8px;
}

.detail-path {
  font-size: 11px;
  color: #8892a6;
  font-family: 'SF Mono', Menlo, monospace;
  word-break: break-all;
  background: #1a1e28;
  padding: 8px 10px;
  border-radius: 6px;
  line-height: 1.5;
}

.detail-content {
  max-height: 280px;
  overflow-y: auto;
  background: #0f1115;
  border: 1px solid #242936;
  border-radius: 6px;
  padding: 12px;
}

.detail-content pre {
  font-family: 'SF Mono', Menlo, monospace;
  font-size: 12px;
  line-height: 1.6;
  color: #c8cfdb;
  white-space: pre-wrap;
  word-break: break-word;
}

.linked-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 200px;
  overflow-y: auto;
}

.linked-item {
  padding: 6px 10px;
  font-size: 12px;
  color: #a8b0c0;
  cursor: pointer;
  border-radius: 4px;
  transition: all 0.1s;
}

.linked-item:hover {
  background: #1d212c;
  color: #e8e8e8;
}
</style>
