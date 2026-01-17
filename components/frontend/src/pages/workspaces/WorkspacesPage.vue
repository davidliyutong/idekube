<template>
  <div class="workspaces-page">
    <div class="page-header">
      <h1 class="page-title">Workspaces</h1>
      <va-button @click="showCreateDialog = true">
        <va-icon name="add" class="mr-2" />
        New Workspace
      </va-button>
    </div>

    <!-- Filters -->
    <va-card class="mb-4">
      <va-card-content>
        <div class="filters">
          <va-input
            v-model="searchQuery"
            placeholder="Search workspaces..."
            class="filter-input"
            @update:model-value="handleSearch"
          >
            <template #prepend>
              <va-icon name="search" />
            </template>
          </va-input>

          <va-select
            v-model="statusFilter"
            label="Status"
            :options="statusOptions"
            class="filter-select"
            @update:model-value="handleFilterChange"
          />

          <va-button preset="secondary" @click="resetFilters">
            <va-icon name="refresh" class="mr-2" />
            Reset
          </va-button>
        </div>
      </va-card-content>
    </va-card>

    <!-- Workspaces Table -->
    <va-card :loading="workspaceStore.loading">
      <va-card-content>
        <va-data-table
          :items="workspaceStore.workspaces"
          :columns="columns"
          :loading="workspaceStore.loading"
          striped
          hoverable
        >
          <template #cell(name)="{ rowData }">
            <router-link
              :to="{ name: 'workspace-detail', params: { id: rowData.id } }"
              class="workspace-link"
            >
              {{ rowData.name }}
            </router-link>
          </template>

          <template #cell(status)="{ rowData }">
            <va-badge :color="getStatusColor(rowData.status)" :text="rowData.status" />
          </template>

          <template #cell(template_name)="{ rowData }">
            {{ rowData.template_name || rowData.template_id }}
          </template>

          <template #cell(owner_username)="{ rowData }">
            {{ rowData.owner_username || rowData.owner_id }}
          </template>

          <template #cell(created_at)="{ rowData }">
            {{ formatDate(rowData.created_at) }}
          </template>

          <template #cell(actions)="{ rowData }">
            <div class="actions">
              <va-button
                v-if="rowData.status === 'stopped'"
                preset="primary"
                size="small"
                @click="handleStart(rowData.id)"
              >
                <va-icon name="play_arrow" />
              </va-button>

              <va-button
                v-if="rowData.status === 'running'"
                preset="warning"
                size="small"
                @click="handleStop(rowData.id)"
              >
                <va-icon name="stop" />
              </va-button>

              <va-button
                preset="secondary"
                size="small"
                @click="handleEdit(rowData)"
              >
                <va-icon name="edit" />
              </va-button>

              <va-button
                preset="danger"
                size="small"
                @click="handleDelete(rowData)"
              >
                <va-icon name="delete" />
              </va-button>
            </div>
          </template>
        </va-data-table>

        <!-- Pagination -->
        <div class="pagination">
          <va-pagination
            v-model="currentPage"
            :pages="totalPages"
            :visible-pages="5"
            @update:model-value="handlePageChange"
          />
          <div class="pagination-info">
            Showing {{ startIndex }} - {{ endIndex }} of {{ workspaceStore.total }} workspaces
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create Dialog Placeholder -->
    <CreateWorkspaceModal
      v-model="showCreateDialog"
      @created="loadWorkspaces"
    />

    <!-- Delete Confirmation -->
    <va-modal
      v-model="showDeleteDialog"
      title="Delete Workspace"
      message="Are you sure you want to delete this workspace? This action cannot be undone."
      ok-text="Delete"
      cancel-text="Cancel"
      @ok="confirmDelete"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useWorkspaceStore } from '@/stores/workspace'
import { formatDate as formatDateUtil } from '@/utils/date'
import CreateWorkspaceModal from '@/components/CreateWorkspaceModal.vue'

const router = useRouter()
const workspaceStore = useWorkspaceStore()

// State
const showCreateDialog = ref(false)
const showDeleteDialog = ref(false)
const workspaceToDelete = ref<any>(null)
const searchQuery = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = 20

// Options
const statusOptions = [
  { text: 'All', value: '' },
  { text: 'Running', value: 'running' },
  { text: 'Stopped', value: 'stopped' },
  { text: 'Pending', value: 'pending' },
  { text: 'Error', value: 'error' },
]

const columns = [
  { key: 'name', label: 'Name', sortable: true },
  { key: 'status', label: 'Status', sortable: true },
  { key: 'template_name', label: 'Template' },
  { key: 'owner_username', label: 'Owner' },
  { key: 'created_at', label: 'Created', sortable: true },
  { key: 'actions', label: 'Actions', width: 200 },
]

// Computed
const totalPages = computed(() => Math.ceil(workspaceStore.total / pageSize))

const startIndex = computed(() => (currentPage.value - 1) * pageSize + 1)

const endIndex = computed(() => {
  const end = currentPage.value * pageSize
  return end > workspaceStore.total ? workspaceStore.total : end
})

// Methods
function formatDate(date: string): string {
  return formatDateUtil(date)
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

async function loadWorkspaces(): Promise<void> {
  await workspaceStore.fetchWorkspaces({
    page: currentPage.value,
    page_size: pageSize,
    search: searchQuery.value || undefined,
    status: statusFilter.value || undefined,
  })
}

function handleSearch(): void {
  currentPage.value = 1
  loadWorkspaces()
}

function handleFilterChange(): void {
  currentPage.value = 1
  loadWorkspaces()
}

function resetFilters(): void {
  searchQuery.value = ''
  statusFilter.value = ''
  currentPage.value = 1
  loadWorkspaces()
}

function handlePageChange(): void {
  loadWorkspaces()
}

async function handleStart(id: string): Promise<void> {
  try {
    await workspaceStore.startWorkspace(id)
  } catch (error) {
    console.error('Failed to start workspace:', error)
  }
}

async function handleStop(id: string): Promise<void> {
  try {
    await workspaceStore.stopWorkspace(id)
  } catch (error) {
    console.error('Failed to stop workspace:', error)
  }
}

function handleEdit(workspace: any): void {
  router.push({ name: 'workspace-detail', params: { id: workspace.id } })
}

function handleDelete(workspace: any): void {
  workspaceToDelete.value = workspace
  showDeleteDialog.value = true
}

async function confirmDelete(): Promise<void> {
  if (workspaceToDelete.value) {
    try {
      await workspaceStore.deleteWorkspace(workspaceToDelete.value.id)
      workspaceToDelete.value = null
    } catch (error) {
      console.error('Failed to delete workspace:', error)
    }
  }
}

onMounted(() => {
  loadWorkspaces()
  workspaceStore.startAutoRefresh()
})

onUnmounted(() => {
  workspaceStore.stopAutoRefresh()
})
</script>

<style lang="scss" scoped>
.workspaces-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .filters {
    display: flex;
    gap: 1rem;
    align-items: flex-end;

    .filter-input {
      flex: 1;
      min-width: 200px;
    }

    .filter-select {
      min-width: 150px;
    }
  }

  .workspace-link {
    color: var(--va-primary);
    text-decoration: none;
    font-weight: 500;

    &:hover {
      text-decoration: underline;
    }
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 1.5rem;

    .pagination-info {
      color: var(--va-text-secondary);
      font-size: 0.875rem;
    }
  }
}
</style>
