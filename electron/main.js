const { app, BrowserWindow, ipcMain, dialog } = require('electron')
const path = require('path')
const { spawn, exec } = require('child_process')
const fs = require('fs')

let mainWindow = null
let goServerProcess = null
const GO_SERVER_PORT = 37245

function findGoServer() {
  const candidates = [
    path.join(__dirname, '..', 'bin', process.platform === 'win32' ? 'md-server.exe' : 'md-server'),
    path.join(process.resourcesPath, 'bin', process.platform === 'win32' ? 'md-server.exe' : 'md-server'),
  ]
  for (const p of candidates) {
    if (fs.existsSync(p)) {
      return p
    }
  }
  return null
}

function startGoServer() {
  if (process.env.NODE_ENV === 'development') {
    console.log('[Electron] Development mode: expecting Go server to run separately on port', GO_SERVER_PORT)
    return
  }
  const serverPath = findGoServer()
  if (!serverPath) {
    console.error('[Electron] Go server binary not found')
    return
  }
  console.log('[Electron] Starting Go server:', serverPath)
  goServerProcess = spawn(serverPath, [], {
    env: { ...process.env, PORT: String(GO_SERVER_PORT) },
    stdio: 'ignore',
  })
  goServerProcess.on('error', (err) => {
    console.error('[Electron] Go server error:', err)
  })
  goServerProcess.on('exit', (code) => {
    console.log('[Electron] Go server exited with code:', code)
  })
}

function stopGoServer() {
  if (goServerProcess) {
    goServerProcess.kill()
    goServerProcess = null
  }
}

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1280,
    height: 860,
    minWidth: 960,
    minHeight: 640,
    title: 'MD Note Graph',
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      contextIsolation: true,
      nodeIntegration: false,
    },
  })

  if (process.env.NODE_ENV === 'development') {
    mainWindow.loadURL('http://localhost:5173')
    mainWindow.webContents.openDevTools({ mode: 'detach' })
  } else {
    mainWindow.loadFile(path.join(__dirname, '..', 'dist', 'index.html'))
  }

  mainWindow.on('closed', () => {
    mainWindow = null
  })
}

ipcMain.handle('select-folder', async () => {
  const result = await dialog.showOpenDialog(mainWindow, {
    properties: ['openDirectory'],
    title: '选择 Markdown 笔记文件夹',
  })
  if (result.canceled || result.filePaths.length === 0) {
    return null
  }
  return result.filePaths[0]
})

ipcMain.handle('get-server-port', () => {
  return GO_SERVER_PORT
})

ipcMain.handle('save-svg-file', async (_event, { content, defaultName }) => {
  if (!content) {
    return { success: false, message: 'SVG 内容为空' }
  }
  const safeName = (defaultName || 'note-graph.svg').replace(/[\\/:*?"<>|]/g, '_')
  const result = await dialog.showSaveDialog(mainWindow, {
    title: '保存 SVG 关系图',
    defaultPath: safeName,
    filters: [{ name: 'SVG 图片', extensions: ['svg'] }],
  })
  if (result.canceled || !result.filePath) {
    return { success: false, canceled: true }
  }
  try {
    fs.writeFileSync(result.filePath, content, 'utf-8')
    return { success: true, path: result.filePath }
  } catch (err) {
    return { success: false, message: String(err && err.message ? err.message : err) }
  }
})

app.whenReady().then(() => {
  startGoServer()
  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow()
    }
  })
})

app.on('window-all-closed', () => {
  stopGoServer()
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('before-quit', () => {
  stopGoServer()
})

process.on('SIGINT', () => {
  stopGoServer()
  process.exit(0)
})
