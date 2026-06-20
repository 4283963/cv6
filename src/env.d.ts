/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

interface ElectronAPI {
  selectFolder: () => Promise<string | null>
  getServerPort: () => Promise<number>
}

declare global {
  interface Window {
    electronAPI?: ElectronAPI
  }
}

export {}
