// Re-export the configured axios client
export { default as apiClient } from './client'

// Re-export generated API clients
// After running `make generate-api`, these exports provide type-safe API access
export * from './client/api'
export * from './client/models'

// Usage examples:
//
// 1. Import API classes:
//    import { DefaultApi, MFAApi, OIDCApi } from '@/api'
//
// 2. Import models/types:
//    import type { ModelsWorkspace, ModelsUser } from '@/api'
//
// 3. Create API instance:
//    import { DefaultApi, apiClient } from '@/api'
//    const api = new DefaultApi(undefined, '', apiClient)
//    const response = await api.workspacesGet()
//
// See docs/API_CLIENT.md for comprehensive usage guide

