<template>
  <div class="workspace-detail-page">
    <div v-if="workspaceStore.loading && !workspace" class="loading-container">
      <va-progress-circle indeterminate />
    </div>

    <div v-else-if="workspace" class="workspace-detail">
      <!-- Header -->
      <div class="page-header">
        <div class="header-left">
          <va-button preset="secondary" size="small" @click="goBack">
            <va-icon name="arrow_back" class="mr-2" />
            Back
          </va-button>
          <h1 class="page-title">{{ workspace.name }}</h1>
          <va-badge :color="getStatusColor(workspace.status)" :text="workspace.status" />
        </div>
        <div class="header-actions">
          <va-button
            v-if="workspace.status === 'stopped'"
            preset="primary"
            @click="handleStart"
            :loading="workspaceStore.loading"
          >
            <va-icon name="play_arrow" class="mr-2" />
            Start
          </va-button>
          <va-button
            v-if="workspace.status === 'running'"
            preset="warning"
            @click="handleStop"
            :loading="workspaceStore.loading"
          >
            <va-icon name="stop" class="mr-2" />
            Stop
          </va-button>
          <va-button preset="secondary" @click="showEditDialog = true">
            <va-icon name="edit" class="mr-2" />
            Edit
          </va-button>
          <va-button preset="danger" @click="showDeleteDialog = true">
            <va-icon name="delete" class="mr-2" />
            Delete
          </va-button>
        </div>
      </div>

      <!-- Info Cards -->
      <div class="info-grid">
        <!-- Basic Info -->
        <va-card>
          <va-card-title>Basic Information</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="info-label">ID:</span>
              <span class="info-value">{{ workspace.id }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Description:</span>
              <span class="info-value">{{ workspace.description || 'N/A' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Template:</span>
              <span class="info-value">{{ workspace.template_name || workspace.template_id }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Owner:</span>
              <span class="info-value">{{ workspace.owner_username || workspace.owner_id }}</span>
            </div>
            <div v-if="workspace.organization_name" class="info-row">
              <span class="info-label">Organization:</span>
              <span class="info-value">{{ workspace.organization_name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Created:</span>
              <span class="info-value">{{ formatDate(workspace.created_at) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Updated:</span>
              <span class="info-value">{{ formatDate(workspace.updated_at) }}</span>
            </div>
          </va-card-content>
        </va-card>

        <!-- Resource Limits -->
        <va-card>
          <va-card-title>Resource Limits</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="info-label">CPU:</span>
              <span class="info-value">{{ workspace.cpu_limit || 'Unlimited' }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Memory:</span>
              <span class="info-value">{{ formatMemory(workspace.memory_limit) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Storage:</span>
              <span class="info-value">{{ formatStorage(workspace.storage_limit) }}</span>
            </div>
            <div v-if="workspace.image" class="info-row">
              <span class="info-label">Image:</span>
              <span class="info-value code">{{ workspace.image }}</span>
            </div>
            <div v-if="workspace.port" class="info-row">
              <span class="info-label">Port:</span>
              <span class="info-value">{{ workspace.port }}</span>
            </div>
          </va-card-content>
        </va-card>
      </div>

      <!-- Environment Variables -->
      <va-card v-if="workspace.env_vars && Object.keys(workspace.env_vars).length > 0" class="mt-4">
        <va-card-title>Environment Variables</va-card-title>
        <va-card-content>
          <va-data-table
            :items="envVarsList"
            :columns="envVarsColumns"
            striped
          />
        </va-card-content>
      </va-card>

      <!-- Attached Volumes -->
      <va-card class="mt-4">
        <va-card-title>
          <div class="card-title-with-action">
            <span>Attached Volumes</span>
            <va-button size="small" @click="showAttachVolumeDialog = true">
              <va-icon name="add" class="mr-2" />
              Attach Volume
            </va-button>
          </div>
        </va-card-title>
        <va-card-content>
          <div v-if="!workspace.volumes || workspace.volumes.length === 0" class="empty-state">
            <p>No volumes attached</p>
          </div>
          <va-data-table
            v-else
            :items="workspace.volumes"
            :columns="volumesColumns"
            striped
          >
            <template #cell(actions)="{ rowData }">
              <va-button
                preset="danger"
                size="small"
                @click="handleDetachVolume(rowData.id)"
              >
                <va-icon name="remove" class="mr-2" />
                Detach
              </va-button>
            </template>
          </va-data-table>
        </va-card-content>
      </va-card>
    </div>

    <div v-else class="error-state">
      <p>Workspace not found</p>
      <va-button @click="goBack">Go Back</va-button>
    </div>

    <!-- Edit Dialog Placeholder -->
    <va-modal v-model="showEditDialog" title="Edit Workspace" size="large">
      <p>Workspace edit form will be implemented here</p>
      <template #footer>
        <va-button preset="secondary" @click="showEditDialog = false">Cancel</va-button>
        <va-button @click="handleUpdate">Save</va-button>
      </template>
    </va-modal>

    <!-- Delete Confirmation -->
    <va-modal
      v-model="showDeleteDialog"
      title="Delete Workspace"
      message="Are you sure you want to delete this workspace? This action cannot be undone."
      ok-text="Delete"
      cancel-text="Cancel"
      @ok="confirmDelete"
    />

    <!-- Attach Volume Dialog Placeholder -->
    <va-modal v-model="showAttachVolumeDialog" title="Attach Volume" size="medium">
      <p>Volume attachment form will be implemented here</p>
      <template #footer>
        <va-button preset="secondary" @click="showAttachVolumeDialog = false">Cancel</va-button>
        <va-button @click="handleAttachVolume">Attach</va-button>
      </template>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatDate as formatDateUtil } from '@/utils/date'

const route = useRoute()
const router = useRouter()
const workspaceStore = useWorkspaceStore()

// State
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const showAttachVolumeDialog = ref(false)

// Computed
const workspaceId = computed(() => route.params.id as string)
const workspace = computed(() => workspaceStore.currentWorkspace)

const envVarsList = computed(() => {
  if (!workspace.value?.env_vars) return []
  return Object.entries(workspace.value.env_vars).map(([key, value]) => ({
    key,
    value,
  }))
})

const envVarsColumns = [
  { key: 'key', label: 'Key' },
  { key: 'value', label: 'Value' },
]

const volumesColumns = [
  { key: 'name', label: 'Name' },
  { key: 'mount_path', label: 'Mount Path' },
  { key: 'actions', label: 'Actions', width: 150 },
]

// Methods
function formatDate(date: string): string {
  return formatDateUtil(date)
}

function formatMemory(bytes: number | undefined): string {
  if (!bytes) return 'Unlimited'
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(2)} GB`
}

function formatStorage(bytes: number | undefined): string {
  if (!bytes) return 'Unlimited'
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(2)} GB`
}

function getStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    running: 'success',
    stopped: 'secondary',
    pending: 'info',
    error: 'danger',
    deleting: 'warning',
  }
  return colorMap[status] || 'secondary'
}

function goBack(): void {
  router.push({ name: 'workspaces' })
}

async function handleStart(): Promise<void> {
  try {
    await workspaceStore.startWorkspace(workspaceId.value)
  } catch (error) {
    console.error('Failed to start workspace:', error)
  }
}

async function handleStop(): Promise<void> {
  try {
    await workspaceStore.stopWorkspace(workspaceId.value)
  } catch (error) {
    console.error('Failed to stop workspace:', error)
  }
}

function handleUpdate(): void {
  // Placeholder
  console.log('Update workspace')
  showEditDialog.value = false
}

async function confirmDelete(): Promise<void> {
  try {
    await workspaceStore.deleteWorkspace(workspaceId.value)
    router.push({ name: 'workspaces' })
  } catch (error) {
    console.error('Failed to delete workspace:', error)
  }
}

function handleAttachVolume(): void {
  // Placeholder
  console.log('Attach volume')
  showAttachVolumeDialog.value = false
}

async function handleDetachVolume(volumeId: string): Promise<void> {
  try {
    await workspaceStore.detachVolume(workspaceId.value, volumeId)
  } catch (error) {
    console.error('Failed to detach volume:', error)
  }
}

async function loadWorkspace(): Promise<void> {
  try {
    await workspaceStore.fetchWorkspace(workspaceId.value)
  } catch (error) {
    console.error('Failed to load workspace:', error)
  }
}

// Lifecycle
onMounted(() => {
  loadWorkspace()
  // Refresh every 10 seconds
  const interval = setInterval(loadWorkspace, 10000)
  onUnmounted(() => clearInterval(interval))
})
</script>

<style lang="scss" scoped>
.workspace-detail-page {
  .loading-container,
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    gap: 1rem;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;

    .header-left {
      display: flex;
      align-items: center;
      gap: 1rem;

      .page-title {
        margin: 0;
        font-size: 2rem;
        font-weight: 600;
      }
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
    }
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--va-background-border);

    &:last-child {
      border-bottom: none;
    }

    .info-label {
      font-weight: 500;
      color: var(--va-text-secondary);
    }

    .info-value {
      text-align: right;
      word-break: break-word;
      max-width: 60%;

      &.code {
        font-family: monospace;
        font-size: 0.875rem;
      }
    }
  }

  .card-title-with-action {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }

  .empty-state {
    text-align: center;
    padding: 2rem;
    color: var(--va-text-secondary);
  }

  .mt-4 {
    margin-top: 1.5rem;
  }
}
</style>
