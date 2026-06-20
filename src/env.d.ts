/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface SaveSVGResult {
  success: boolean
  canceled?: boolean
  message?: string
  path?: string
}

interface ElectronAPI {
  selectFolder: () => Promise<string | null>
  getServerPort: () => Promise<number>
  saveSVGFile: (content: string, defaultName: string) => Promise<SaveSVGResult>
}

declare global {
  interface Window {
    electronAPI?: ElectronAPI
  }
}

export {}
