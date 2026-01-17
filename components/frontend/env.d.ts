/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_APP_TITLE: string
  readonly VITE_ENABLE_MOCK: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

// Type augmentations for Vuestic UI
import type { DataTableItem } from 'vuestic-ui'

declare module 'vuestic-ui' {
  // Allow arbitrary property access on DataTableRow
  interface DataTableRow<Item extends DataTableItem = any> {
    [key: string]: any
    [key: number]: any
  }
}
