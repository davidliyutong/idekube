<template>
  <div class="api-keys-page">
    <div class="page-header">
      <h1 class="page-title">API Keys</h1>
      <va-button @click="showCreateModal = true">
        <va-icon name="add" class="mr-2" />
        Create API Key
      </va-button>
    </div>

    <!-- Stats Cards -->
    <div class="stats-grid">
      <va-card class="stat-card">
        <va-card-content>
          <div class="stat-content">
            <va-icon name="key" size="large" color="primary" />
            <div class="stat-info">
              <h3>{{ activeKeys.length }}</h3>
              <p>Active Keys</p>
            </div>
          </div>
        </va-card-content>
      </va-card>

      <va-card class="stat-card">
        <va-card-content>
          <div class="stat-content">
            <va-icon name="schedule" size="large" color="warning" />
            <div class="stat-info">
              <h3>{{ expiredKeys.length }}</h3>
              <p>Expired Keys</p>
            </div>
          </div>
        </va-card-content>
      </va-card>
    </div>

    <!-- API Keys Table -->
    <va-card>
      <va-card-content>
        <va-data-table
          :items="apiKeys"
          :columns="columns"
          :loading="loading"
          striped
          hoverable
        >
          <template #cell(name)="{ row }">
            <strong>{{ row.name }}</strong>
          </template>

          <template #cell(scopes)="{ row }">
            <div class="scopes-cell">
              <va-chip
                v-for="scope in row.scopes.slice(0, 3)"
                :key="scope"
                size="small"
                class="scope-chip"
              >
                {{ scope }}
              </va-chip>
              <va-chip v-if="row.scopes.length > 3" size="small" color="secondary">
                +{{ row.scopes.length - 3 }} more
              </va-chip>
            </div>
          </template>

          <template #cell(status)="{ row }">
            <va-badge
              :color="getStatusColor(row as unknown as APIKey)"
              :text="getStatusText(row as unknown as APIKey)"
            />
          </template>

          <template #cell(last_used_at)="{ row }">
            <span>{{ row.last_used_at ? formatDate(row.last_used_at) : 'Never' }}</span>
          </template>

          <template #cell(created_at)="{ row }">
            <span>{{ formatDate(row.created_at) }}</span>
          </template>

          <template #cell(actions)="{ row }">
            <div class="action-buttons">
              <va-button
                size="small"
                preset="secondary"
                @click="viewDetails(row as unknown as APIKey)"
              >
                <va-icon name="visibility" />
              </va-button>
              <va-button
                size="small"
                color="danger"
                preset="secondary"
                @click="handleDelete(row as unknown as APIKey)"
                :disabled="loading"
              >
                <va-icon name="delete" />
              </va-button>
            </div>
          </template>
        </va-data-table>
      </va-card-content>
    </va-card>

    <!-- Create API Key Modal -->
    <va-modal
      v-model="showCreateModal"
      title="Create API Key"
      size="medium"
      okText="Create"
      @ok="handleCreate"
    >
      <div class="create-form">
        <va-input
          v-model="formData.name"
          label="Key Name"
          placeholder="My API Key"
        />

        <va-select
          v-model="formData.scopes"
          label="Scopes"
          :options="availableScopes"
          multiple
          searchable
        />

        <va-input
          v-model.number="formData.expires_days"
          label="Expires In (days)"
          type="number"
          placeholder="30"
        />
      </div>
    </va-modal>

    <!-- API Key Created Modal -->
    <va-modal
      v-model="showKeyModal"
      title="API Key Created"
      size="medium"
      hide-default-actions
    >
      <div class="key-display">
        <va-alert color="warning" class="mb-4">
          <template #title>Important!</template>
          Make sure to copy your API key now. You won't be able to see it again!
        </va-alert>

        <div class="key-box">
          <code>{{ currentKey }}</code>
          <va-button size="small" @click="copyKey" class="copy-btn">
            <va-icon name="content_copy" />
          </va-button>
        </div>

        <va-button @click="showKeyModal = false" class="mt-4">
          Close
        </va-button>
      </div>
    </va-modal>

    <!-- Details Modal -->
    <va-modal
      v-model="showDetailsModal"
      :title="`API Key: ${selectedKey?.name}`"
      size="medium"
      okText="Close"
      @ok="showDetailsModal = false"
    >
      <div v-if="selectedKey" class="details-content">
        <div class="detail-row">
          <label>Name:</label>
          <span>{{ selectedKey.name }}</span>
        </div>
        <div class="detail-row">
          <label>Status:</label>
          <va-badge
            :color="getStatusColor(selectedKey)"
            :text="getStatusText(selectedKey)"
          />
        </div>
        <div class="detail-row">
          <label>Scopes:</label>
          <div class="scopes-list">
            <va-chip v-for="scope in selectedKey.scopes" :key="scope" size="small">
              {{ scope }}
            </va-chip>
          </div>
        </div>
        <div class="detail-row">
          <label>Created:</label>
          <span>{{ formatDate(selectedKey.created_at) }}</span>
        </div>
        <div class="detail-row">
          <label>Expires:</label>
          <span>{{ selectedKey.expires_at ? formatDate(selectedKey.expires_at) : 'Never' }}</span>
        </div>
        <div class="detail-row">
          <label>Last Used:</label>
          <span>{{ selectedKey.last_used_at ? formatDate(selectedKey.last_used_at) : 'Never' }}</span>
        </div>
      </div>
    </va-modal>

    <!-- Delete Confirmation -->
    <va-modal
      v-model="showDeleteConfirm"
      title="Confirm Delete"
      okText="Delete"
      cancelText="Cancel"
      @ok="confirmDelete"
    >
      <p>Are you sure you want to delete API key <strong>{{ keyToDelete?.name }}</strong>?</p>
      <p class="text-secondary">This action cannot be undone.</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAPIKeyStore, type APIKey } from '@/stores/apiKey'
import { formatDate } from '@/utils/date'

const apiKeyStore = useAPIKeyStore()

const showCreateModal = ref(false)
const showKeyModal = ref(false)
const showDetailsModal = ref(false)
const showDeleteConfirm = ref(false)
const selectedKey = ref<APIKey | null>(null)
const keyToDelete = ref<APIKey | null>(null)

const formData = ref({
  name: '',
  scopes: [] as string[],
  expires_days: undefined as number | undefined,
})

const availableScopes = [
  'workspaces:read', 'workspaces:write',
  'templates:read', 'templates:write',
  'volumes:read', 'volumes:write',
  'users:read', 'organizations:read',
]

const { apiKeys, currentKey, loading, activeKeys, expiredKeys } = apiKeyStore

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'scopes', label: 'Scopes' },
  { key: 'status', label: 'Status' },
  { key: 'last_used_at', label: 'Last Used' },
  { key: 'created_at', label: 'Created' },
  { key: 'actions', label: 'Actions', width: '120px' },
]

function getStatusColor(key: APIKey): string {
  if (!key.expires_at) return 'success'
  return new Date(key.expires_at) > new Date() ? 'success' : 'danger'
}

function getStatusText(key: APIKey): string {
  if (!key.expires_at) return 'Active'
  return new Date(key.expires_at) > new Date() ? 'Active' : 'Expired'
}

async function handleCreate() {
  try {
    const data: any = {
      name: formData.value.name,
      scopes: formData.value.scopes,
    }
    if (formData.value.expires_days) {
      const expiresAt = new Date()
      expiresAt.setDate(expiresAt.getDate() + formData.value.expires_days)
      data.expires_at = Math.floor(expiresAt.getTime() / 1000)
    }
    await apiKeyStore.createAPIKey(data)
    showCreateModal.value = false
    showKeyModal.value = true
    formData.value = { name: '', scopes: [], expires_days: undefined }
  } catch (error) {
    console.error('Failed to create API key:', error)
  }
}

function copyKey() {
  if (currentKey) navigator.clipboard.writeText(currentKey)
}

function viewDetails(key: APIKey) {
  selectedKey.value = key
  showDetailsModal.value = true
}

function handleDelete(key: APIKey) {
  keyToDelete.value = key
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (keyToDelete.value) {
    await apiKeyStore.deleteAPIKey(keyToDelete.value.id)
    showDeleteConfirm.value = false
    keyToDelete.value = null
  }
}

onMounted(() => {
  apiKeyStore.fetchAPIKeys()
})
</script>

<style lang="scss" scoped>
.api-keys-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;

    .stat-card .stat-content {
      display: flex;
      align-items: center;
      gap: 1rem;

      .stat-info h3 {
        margin: 0;
        font-size: 2rem;
        font-weight: 600;
      }

      .stat-info p {
        margin: 0;
        color: var(--va-textColorSecondary);
        font-size: 0.875rem;
      }
    }
  }

  .scopes-cell {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
  }
}

.create-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.key-display .key-box {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background-color: var(--va-background-shade);
  border-radius: 0.5rem;
  border: 1px solid var(--va-background-border);

  code {
    flex: 1;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 0.875rem;
    word-break: break-all;
  }
}

.details-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;

  .detail-row {
    display: flex;
    align-items: flex-start;
    gap: 1rem;

    label {
      min-width: 100px;
      font-weight: 500;
      color: var(--va-textColorSecondary);
    }

    span {
      flex: 1;
    }

    .scopes-list {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }
  }
}

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-2 { margin-right: 0.5rem; }
.mb-4 { margin-bottom: 1rem; }
.mt-4 { margin-top: 1rem; }
</style>
