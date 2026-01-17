// Global type augmentations
declare global {
  // Type helper for DataTable rows - cast to the original type
  type UnwrapDataTableRow<T> = T extends import('vuestic-ui').DataTableRow<infer U> ? U : T
}

// Extend Vuestic UI module
declare module 'vuestic-ui' {
  // Override DataTableRow to be transparent (pass-through type)
  export type DataTableRow<T> = T
  
  export interface ValidationRule<T = any> {
    (v: T): boolean | string
  }
}

export {}
