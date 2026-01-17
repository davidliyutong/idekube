<template>
  <div class="volumes-page">
    <div class="page-header">
      <h1 class="page-title">数据卷管理</h1>
      <va-button @click="showCreateDialog = true">
        <va-icon name="add" class="mr-2" />
        新建数据卷
      </va-button>
    </div>

    <!-- Filters -->
    <va-card class="filters-card">
      <va-card-content>
        <div class="filters-row">
          <va-input
            v-model="searchQuery"
            placeholder="搜索数据卷名称..."
            class="filter-item"
            clearable
          >
            <template #prependInner>
              <va-icon name="search" />
            </template>
          </va-input>

          <va-select
            v-model="selectedStatus"
            :options="statusOptions"
            label="状态筛选"
            class="filter-item"
            clearable
          />

          <va-button @click="fetchVolumes" preset="secondary">
            <va-icon name="refresh" class="mr-2" />
            刷新
          </va-button>
        </div>
      </va-card-content>
    </va-card>

    <!-- Data Table -->
    <va-card class="data-card">
      <va-card-content>
        <div class="table-wrapper">
          <va-data-table
            :items="filteredVolumes"
            :columns="columns"
            striped
            hoverable
            :loading="loading"
          >
            <template #cell(name)="{ row }">
              <router-link :to="`/app/volumes/${row.id}`" class="link-cell">
                <strong>{{ row.name }}</strong>
              </router-link>
            </template>

            <template #cell(size)="{ row }">
              <span>{{ formatSize(row.size) }}</span>
            </template>

            <template #cell(attached_to)="{ row }">
              <div class="attached-list">
                <template v-if="row.attached_to && row.attached_to.length > 0">
                  <va-chip
                    v-for="workspaceId in row.attached_to.slice(0, 2)"
                    :key="workspaceId"
                    size="small"
                    class="mb-2"
                  >
                    {{ workspaceId }}
                  </va-chip>
                  <template v-if="row.attached_to.length > 2">
                    <va-chip size="small" class="mb-2">
                      +{{ row.attached_to.length - 2 }} 更多
                    </va-chip>
                  </template>
                </template>
                <span v-else class="text-secondary">未挂载</span>
              </div>
            </template>

            <template #cell(status)="{ row }">
              <va-badge
                :color="getStatusColor(row.status)"
                :text-color="getStatusColor(row.status) === 'warning' ? 'black' : 'white'"
              >
                {{ getStatusLabel(row.status) }}
              </va-badge>
            </template>

            <template #cell(created_at)="{ row }">
              <span>{{ formatDate(row.created_at) }}</span>
            </template>

            <template #cell(actions)="{ row }">
              <div class="action-buttons">
                <va-button
                  size="small"
                  preset="secondary"
                  @click="handleSync(row.id)"
                  :disabled="loading"
                >
                  <va-icon name="sync" />
                </va-button>
                <va-button
                  size="small"
                  preset="secondary"
                  @click="editVolume(row as any)"
                  :disabled="loading"
                >
                  <va-icon name="edit" />
                </va-button>
                <va-button
                  size="small"
                  color="danger"
                  preset="secondary"
                  @click="handleDelete(row.id)"
                  :disabled="loading"
                >
                  <va-icon name="delete" />
                </va-button>
              </div>
            </template>
          </va-data-table>
        </div>

        <!-- Pagination -->
        <div class="pagination-wrapper">
          <va-pagination
            v-model="currentPage"
            :pages="totalPages"
            @update:model-value="onPageChange"
          />
          <span class="pagination-info">
            共 {{ total }} 个数据卷
          </span>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create/Edit Modal -->
    <create-volume-modal
      v-model="showCreateDialog"
      :editing-volume="editingVolume"
      @created="handleVolumeCreated"
      @updated="handleVolumeUpdated"
    />

    <!-- Delete Confirmation Modal -->
    <va-modal
      v-model="showDeleteConfirm"
      title="确认删除"
      okText="删除"
      cancelText="取消"
      @ok="confirmDelete"
    >
      <p>确定要删除数据卷 <strong>{{ volumeToDelete?.name }}</strong> 吗？</p>
      <p class="text-secondary">此操作不可逆，请谨慎操作。</p>
    </va-modal>

    <!-- Sync Progress Modal -->
    <va-modal
      v-model="showSyncProgress"
      title="数据卷同步"
      :closable="!isSyncing"
      :ok-text="isSyncing ? '同步中...' : '关闭'"
      @ok="showSyncProgress = false"
    >
      <div class="sync-content">
        <va-progress
          :model-value="syncProgress"
          :color="syncProgress === 100 ? 'success' : 'info'"
        />
        <p class="mt-3">{{ syncStatusMessage }}</p>
      </div>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useVolumeStore, type Volume } from '@/stores/volume'
import CreateVolumeModal from '@/components/CreateVolumeModal.vue'
import { formatDate } from '@/utils/date'

const volumeStore = useVolumeStore()

// Ref
const searchQuery = ref('')
const selectedStatus = ref<string | null>(null)
const showCreateDialog = ref(false)
const editingVolume = ref<Volume | null>(null)
const showDeleteConfirm = ref(false)
const volumeToDelete = ref<Volume | null>(null)
const showSyncProgress = ref(false)
const syncProgress = ref(0)
const isSyncing = ref(false)
let syncInterval: ReturnType<typeof setInterval> | null = null

// Table columns
const columns: any[] = [
  { key: 'name', label: '名称', width: '200px' },
  { key: 'description', label: '描述', width: '250px' },
  { key: 'size', label: '大小' },
  { key: 'attached_to', label: '已挂载工作空间', width: '300px' },
  { key: 'status', label: '状态' },
  { key: 'created_at', label: '创建时间' },
  { key: 'actions', label: '操作', width: '150px', align: 'center' },
]

const statusOptions = [
  { text: '可用', value: 'available' },
  { text: '使用中', value: 'in-use' },
  { text: '错误', value: 'error' },
]

// Computed
const { volumes, loading, total, currentPage, pageSize } = volumeStore

const totalPages = computed(() => Math.ceil(total / pageSize) || 1)

const filteredVolumes = computed(() => {
  let result = volumes

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(
      (v) => v.name.toLowerCase().includes(query) || v.description?.toLowerCase().includes(query)
    )
  }

  if (selectedStatus.value) {
    result = result.filter((v) => v.status === selectedStatus.value)
  }

  return result
})

const syncStatusMessage = computed(() => {
  if (syncProgress.value === 0) return '准备同步...'
  if (syncProgress.value < 100) return `同步中... ${syncProgress.value}%`
  return '同步完成'
})

// Methods
function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`
}

function getStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    available: 'success',
    'in-use': 'info',
    error: 'danger',
  }
  return colorMap[status] || 'secondary'
}

function getStatusLabel(status: string): string {
  const labelMap: Record<string, string> = {
    available: '可用',
    'in-use': '使用中',
    error: '错误',
  }
  return labelMap[status] || status
}

async function fetchVolumes() {
  await volumeStore.fetchVolumes({
    page: currentPage,
    page_size: pageSize,
  })
}

function editVolume(volume: Volume) {
  editingVolume.value = volume
  showCreateDialog.value = true
}

function handleDelete(volumeId: string) {
  volumeToDelete.value = volumes.find((v) => v.id === volumeId) || null
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (volumeToDelete.value) {
    await volumeStore.deleteVolume(volumeToDelete.value.id)
    showDeleteConfirm.value = false
    volumeToDelete.value = null
    await fetchVolumes()
  }
}

async function handleSync(volumeId: string) {
  showSyncProgress.value = true
  syncProgress.value = 0
  isSyncing.value = true

  // Simulate sync progress
  syncInterval = setInterval(() => {
    syncProgress.value = Math.min(syncProgress.value + Math.random() * 30, 95)
  }, 500)

  try {
    await volumeStore.syncVolume(volumeId)
    syncProgress.value = 100
    isSyncing.value = false
  } catch (error) {
    isSyncing.value = false
    syncProgress.value = 0
  } finally {
    if (syncInterval) {
      clearInterval(syncInterval)
      syncInterval = null
    }
  }
}

function handleVolumeCreated() {
  editingVolume.value = null
  showCreateDialog.value = false
  fetchVolumes()
}

function handleVolumeUpdated() {
  editingVolume.value = null
  showCreateDialog.value = false
  fetchVolumes()
}

function onPageChange(page: number) {
  volumeStore.currentPage = page
  fetchVolumes()
}

// Lifecycle
onMounted(() => {
  fetchVolumes()
})

onUnmounted(() => {
  if (syncInterval) {
    clearInterval(syncInterval)
  }
})
</script>

<style lang="scss" scoped>
.volumes-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
  }

  .page-title {
    margin: 0;
    font-size: 2rem;
    font-weight: 600;
  }

  .filters-card {
    margin-bottom: 1.5rem;
  }

  .filters-row {
    display: flex;
    gap: 1rem;
    align-items: flex-end;
    flex-wrap: wrap;

    .filter-item {
      flex: 1;
      min-width: 200px;
    }
  }

  .data-card {
    .table-wrapper {
      overflow-x: auto;
      margin-bottom: 1rem;
    }

    .link-cell {
      color: var(--va-background-border);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }

    .attached-list {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .action-buttons {
      display: flex;
      gap: 0.5rem;
      justify-content: center;
    }

    .pagination-wrapper {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-top: 1rem;
      padding-top: 1rem;
      border-top: 1px solid var(--va-background-border);

      .pagination-info {
        color: var(--va-textColorSecondary);
        font-size: 0.875rem;
      }
    }
  }

  .sync-content {
    padding: 1rem 0;
  }

  .text-secondary {
    color: var(--va-textColorSecondary);
  }

  .mb-2 {
    margin-bottom: 0.5rem;
  }

  .mr-2 {
    margin-right: 0.5rem;
  }

  .mt-3 {
    margin-top: 1rem;
  }
}
</style>
