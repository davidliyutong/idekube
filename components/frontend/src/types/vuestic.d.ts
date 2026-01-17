// Type declarations for Vuestic UI Component Library
// This file augments the 'vuestic-ui' module to add index signatures to DataTableRow

import type { DataTableItem } from 'vuestic-ui'

declare module 'vuestic-ui' {
  // Augment the existing DataTableRow interface to allow arbitrary property access
  // This enables row.property access in templates without TypeScript errors
  interface DataTableRow<Item extends DataTableItem = DataTableItem> {
    // Index signature to allow arbitrary property access
    [key: string]: any
    [key: number]: any
  }
  
  // Add ValidationRule interface for form validation
  interface ValidationRule<T = any> {
    (v: T): boolean | string
  }
}

// Global type helpers for component slot parameters  
declare global {
  type AnyRow = Record<string, any>
}

export {}
