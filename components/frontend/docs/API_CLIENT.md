# API Client Usage Guide

This guide explains how to use the auto-generated TypeScript API client in the IDEKube frontend.

## Overview

The API client is automatically generated from the controller's OpenAPI specification (swagger.json) using OpenAPI Generator. It provides type-safe TypeScript interfaces and methods for all backend API endpoints.

## Generating the Client

### Prerequisites

- **Java 8+** - Required by OpenAPI Generator
- **swagger.json** - Must exist at `../controller/docs/api/swagger.json`

### Generate Command

```bash
make generate-api
```

This will:

1. Validate that swagger.json exists
2. Run OpenAPI Generator
3. Create TypeScript client code in `src/api/client/`

### What Gets Generated

```
src/api/client/
├── api/              # API endpoint classes
│   ├── default-api.ts
│   ├── apiapi.ts
│   ├── mfaapi.ts
│   ├── oidcapi.ts
│   ├── permissions-api.ts
│   └── webhook-api.ts
├── models/           # TypeScript interfaces for all models
│   ├── index.ts
│   ├── workspace.ts
│   ├── user.ts
│   ├── template.ts
│   └── ... (85+ model files)
├── base.ts           # Base API class
├── common.ts         # Common types and utilities
├── configuration.ts  # Configuration interface
└── index.ts          # Main export file
```

## Using the Generated Client

### Import API Classes

```typescript
import { DefaultApi, MFAApi, OIDCApi } from '@/api/client';
```

### Import Models (Types)

```typescript
import type { 
  ModelsWorkspace, 
  ModelsUser, 
  ModelsTemplate 
} from '@/api/client';
```

### Creating API Instances

The generated client provides separate API classes for different endpoint groups:

```typescript
import { DefaultApi } from '@/api/client';
import { apiClient } from '@/api/client';

// Use the configured axios instance from our client.ts
const api = new DefaultApi(undefined, '', apiClient);
```

### Example Usage in a Pinia Store

```typescript
// stores/workspace.ts
import { defineStore } from 'pinia';
import { DefaultApi } from '@/api/client';
import type { ModelsWorkspace } from '@/api/client';
import { apiClient } from '@/api/client';

export const useWorkspaceStore = defineStore('workspace', {
  state: () => ({
    workspaces: [] as ModelsWorkspace[],
    loading: false,
    error: null as string | null,
  }),

  actions: {
    async fetchWorkspaces() {
      this.loading = true;
      this.error = null;

      try {
        const api = new DefaultApi(undefined, '', apiClient);
        const response = await api.workspacesGet();
        this.workspaces = response.data.workspaces || [];
      } catch (error) {
        this.error = 'Failed to fetch workspaces';
        throw error;
      } finally {
        this.loading = false;
      }
    },

    async createWorkspace(data: {
      name: string;
      template_id: string;
      cpu_cores?: number;
      memory_gb?: number;
    }) {
      const api = new DefaultApi(undefined, '', apiClient);
      const response = await api.workspacesPost(data);
      return response.data.workspace;
    },

    async deleteWorkspace(id: string) {
      const api = new DefaultApi(undefined, '', apiClient);
      await api.workspacesIdDelete(id);
      // Remove from local state
      this.workspaces = this.workspaces.filter(w => w.id !== id);
    },
  },
});
```

### Example Usage in a Component

```vue
<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { DefaultApi } from '@/api/client';
import type { ModelsWorkspace } from '@/api/client';
import { apiClient } from '@/api/client';
import { handleApiError } from '@/utils/error';

const workspaces = ref<ModelsWorkspace[]>([]);
const loading = ref(false);

const api = new DefaultApi(undefined, '', apiClient);

async function loadWorkspaces() {
  loading.value = true;
  try {
    const response = await api.workspacesGet();
    workspaces.value = response.data.workspaces || [];
  } catch (error) {
    handleApiError(error, 'Failed to load workspaces');
  } finally {
    loading.value = false;
  }
}

onMounted(() => {
  loadWorkspaces();
});
</script>
```

## Available API Classes

The generated client provides these API classes:

- **DefaultApi** - Main API endpoints (workspaces, templates, volumes, etc.)
- **APIApi** - API key management
- **MFAApi** - Multi-factor authentication
- **OIDCApi** - OpenID Connect authentication
- **PermissionsApi** - Permission checking and policy management
- **WebhookApi** - Webhook management

## Type Definitions

All request/response types are available in the `models/` directory:

```typescript
import type {
  ModelsWorkspace,
  ModelsUser,
  ModelsTemplate,
  ModelsVolume,
  ModelsOrganization,
  GithubComDavidliyutongIdekubeControllerInternalHandlersLoginRequest,
  // ... and many more
} from '@/api/client';
```

## Authentication

The generated client automatically uses the axios instance from `src/api/client.ts`, which:

1. Adds JWT token to all requests via interceptor
2. Handles token refresh on 401 responses
3. Manages authentication state via Pinia store

No additional authentication setup is needed when using the generated client.

## Error Handling

Use the `handleApiError` utility for consistent error handling:

```typescript
import { handleApiError } from '@/utils/error';

try {
  await api.workspacesPost(data);
} catch (error) {
  handleApiError(error, 'Failed to create workspace');
}
```

## Regenerating After API Changes

Whenever the backend API changes:

1. Regenerate swagger.json in controller: `cd ../controller && make docs`
2. Regenerate frontend client: `make generate-api`
3. Update your code to use new types/endpoints

## Best Practices

### 1. Use the Configured axios Instance

Always pass `apiClient` from `src/api/client.ts`:

```typescript
import { apiClient } from '@/api/client';
const api = new DefaultApi(undefined, '', apiClient);
```

### 2. Type Your Data

Use generated model types for type safety:

```typescript
const workspace: ModelsWorkspace = response.data.workspace;
```

### 3. Centralize API Calls

Keep API calls in Pinia stores rather than components:

```typescript
// ✅ Good - in store
const useWorkspaceStore = defineStore('workspace', {
  actions: {
    async fetchWorkspaces() {
      const api = new DefaultApi(undefined, '', apiClient);
      // ...
    }
  }
});

// ❌ Avoid - directly in component
const component = {
  async mounted() {
    const api = new DefaultApi(undefined, '', apiClient);
    // ...
  }
}
```

### 4. Handle Errors Consistently

Use the error handling utilities:

```typescript
import { handleApiError, formatErrorMessage } from '@/utils/error';

try {
  await api.someOperation();
} catch (error) {
  handleApiError(error, 'Operation failed');
  // or
  const message = formatErrorMessage(error);
  console.error(message);
}
```

## Troubleshooting

### "Java not found" Error

Install Java:

```bash
# macOS
brew install openjdk

# Or use Azul Zulu JDK
brew install --cask zulu
```

### "swagger.json not found" Error

Generate the controller's API documentation first:

```bash
cd ../controller
make docs
```

### jenv Path Issues

The generate script automatically handles jenv issues by using `/usr/libexec/java_home` on macOS.

### Type Errors After Regeneration

After regenerating the client:

1. Restart your IDE/editor
2. Run `yarn type-check` to see specific errors
3. Update your code to match new types

## Further Reading

- [OpenAPI Generator Documentation](https://openapi-generator.tech/)
- [TypeScript-Axios Generator](https://openapi-generator.tech/docs/generators/typescript-axios/)
- [Controller API Documentation](../controller/docs/API.md)
