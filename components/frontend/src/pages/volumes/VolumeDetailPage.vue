<template>
  <div class="volume-detail-page">
    <div class="page-header">
      <router-link to="/app/volumes" class="back-link">
        <va-icon name="arrow_back" class="mr-2" />
        返回列表
      </router-link>
      <h1 class="page-title">{{ volume?.name || '加载中...' }}</h1>
      <div class="header-actions">
        <va-button
          @click="showEditDialog = true"
          preset="secondary"
          :disabled="loading"
        >
          <va-icon name="edit" class="mr-2" />
          编辑
        </va-button>
        <va-button
          @click="handleSync"
          preset="secondary"
          :disabled="loading || !volume"
        >
          <va-icon name="sync" class="mr-2" />
          同步
        </va-button>
        <va-button
          color="danger"
          preset="secondary"
          @click="handleDelete"
          :disabled="loading || !volume"
        >
          <va-icon name="delete" class="mr-2" />
          删除
        </va-button>
      </div>
    </div>

    <va-progress v-if="loading" indeterminate class="mb-3" />

    <template v-if="volume">
      <!-- Info Cards -->
      <div class="cards-grid">
        <!-- Basic Info -->
        <va-card class="info-card">
          <va-card-title>基本信息</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="label">数据卷ID</span>
              <span class="value">{{ volume.id }}</span>
            </div>
            <div class="info-row">
              <span class="label">名称</span>
              <span class="value">{{ volume.name }}</span>
            </div>
            <div class="info-row">
              <span class="label">描述</span>
              <span class="value">{{ volume.description || '无' }}</span>
            </div>
            <div class="info-row">
              <span class="label">所有者</span>
              <span class="value">{{ volume.owner_username || volume.owner_id }}</span>
            </div>
            <template v-if="volume.organization_name">
              <div class="info-row">
                <span class="label">组织</span>
                <span class="value">{{ volume.organization_name }}</span>
              </div>
            </template>
            <div class="info-row">
              <span class="label">状态</span>
              <va-badge
                :color="getStatusColor(volume.status)"
                :text-color="getStatusColor(volume.status) === 'warning' ? 'black' : 'white'"
              >
                {{ getStatusLabel(volume.status) }}
              </va-badge>
            </div>
          </va-card-content>
        </va-card>

        <!-- Storage Info -->
        <va-card class="info-card">
          <va-card-title>存储信息</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="label">总大小</span>
              <span class="value">{{ formatSize(volume.size) }}</span>
            </div>
            <div class="info-row">
              <span class="label">创建时间</span>
              <span class="value">{{ formatDate(volume.created_at) }}</span>
            </div>
            <div class="info-row">
              <span class="label">更新时间</span>
              <span class="value">{{ formatDate(volume.updated_at) }}</span>
            </div>
          </va-card-content>
        </va-card>

        <!-- Stats -->
        <va-card class="info-card">
          <va-card-title>使用统计</va-card-title>
          <va-card-content>
            <div class="stat-row">
              <span class="label">已挂载工作空间</span>
              <span class="value">
                {{ volume.attached_to?.length || 0 }} 个
              </span>
            </div>
            <template v-if="volume.labels && Object.keys(volume.labels).length > 0">
              <div class="stat-row">
                <span class="label">标签数量</span>
                <span class="value">{{ Object.keys(volume.labels).length }} 个</span>
              </div>
            </template>
          </va-card-content>
        </va-card>
      </div>

      <!-- Attached Workspaces -->
      <va-card v-if="volume.attached_to && volume.attached_to.length > 0" class="section-card">
        <va-card-title>已挂载工作空间 ({{ volume.attached_to.length }})</va-card-title>
        <va-card-content>
          <div class="attached-workspaces">
            <div
              v-for="workspaceId in volume.attached_to"
              :key="workspaceId"
              class="workspace-item"
            >
              <va-icon name="folder" class="mr-2" />
              <router-link :to="`/app/workspaces/${workspaceId}`" class="workspace-link">
                {{ workspaceId }}
              </router-link>
            </div>
          </div>
        </va-card-content>
      </va-card>

      <!-- Labels -->
      <va-card v-if="volume.labels && Object.keys(volume.labels).length > 0" class="section-card">
        <va-card-title>标签</va-card-title>
        <va-card-content>
          <div class="labels-grid">
            <div v-for="(value, key) in volume.labels" :key="key" class="label-item">
              <strong>{{ key }}</strong>
              <span class="text-secondary">{{ value }}</span>
            </div>
          </div>
        </va-card-content>
      </va-card>
    </template>

    <!-- Edit Modal -->
    <create-volume-modal
      v-model="showEditDialog"
      :editing-volume="volume"
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
      <p>确定要删除数据卷 <strong>{{ volume?.name }}</strong> 吗？</p>
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
import { useRouter, useRoute } from 'vue-router'
import { useVolumeStore } from '@/stores/volume'
import CreateVolumeModal from '@/components/CreateVolumeModal.vue'
import { formatDate } from '@/utils/date'

const router = useRouter()
const route = useRoute()
const volumeStore = useVolumeStore()

const id = computed(() => route.params.id as string)

// Refs
const showEditDialog = ref(false)
const showDeleteConfirm = ref(false)
const showSyncProgress = ref(false)
const syncProgress = ref(0)
const isSyncing = ref(false)
let syncInterval: ReturnType<typeof setInterval> | null = null

// Computed
const volume = computed(() => volumeStore.currentVolume)
const loading = computed(() => volumeStore.loading)

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

async function fetchVolume() {
  try {
    await volumeStore.fetchVolume(id.value)
  } catch (error) {
    console.error('Failed to fetch volume:', error)
  }
}

async function handleSync() {
  showSyncProgress.value = true
  syncProgress.value = 0
  isSyncing.value = true

  // Simulate sync progress
  syncInterval = setInterval(() => {
    syncProgress.value = Math.min(syncProgress.value + Math.random() * 30, 95)
  }, 500)

  try {
    await volumeStore.syncVolume(id.value)
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

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  try {
    await volumeStore.deleteVolume(id.value)
    await router.push('/app/volumes')
  } catch (error) {
    console.error('Failed to delete volume:', error)
  }
}

function handleVolumeUpdated() {
  showEditDialog.value = false
  fetchVolume()
}

// Lifecycle
onMounted(() => {
  fetchVolume()
})

onUnmounted(() => {
  if (syncInterval) {
    clearInterval(syncInterval)
  }
})
</script>

<style lang="scss" scoped>
.volume-detail-page {
  .page-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 2rem;

    .back-link {
      color: var(--va-background-border);
      text-decoration: none;
      display: flex;
      align-items: center;
      font-size: 0.875rem;

      &:hover {
        text-decoration: underline;
      }
    }

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
      flex: 1;
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
    }
  }

  .mb-3 {
    margin-bottom: 1rem;
  }

  .cards-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 1.5rem;
    margin-bottom: 1.5rem;

    .info-card {
      .va-card__title {
        font-size: 1.125rem;
        font-weight: 600;
        margin-bottom: 1rem;
      }

      .info-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem 0;
        border-bottom: 1px solid var(--va-background-border);

        &:last-child {
          border-bottom: none;
        }

        .label {
          color: var(--va-textColorSecondary);
          font-weight: 500;
        }

        .value {
          font-weight: 600;
        }
      }

      .stat-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 0.75rem 0;

        .label {
          color: var(--va-textColorSecondary);
          font-weight: 500;
        }

        .value {
          font-weight: 600;
          color: var(--va-background-border);
        }
      }
    }
  }

  .section-card {
    margin-bottom: 1.5rem;

    .va-card__title {
      font-size: 1.125rem;
      font-weight: 600;
      margin-bottom: 1rem;
    }
  }

  .attached-workspaces {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    .workspace-item {
      display: flex;
      align-items: center;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;

      .workspace-link {
        color: var(--va-background-border);
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }
    }
  }

  .labels-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;

    .label-item {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;

      strong {
        color: var(--va-textColor);
      }

      .text-secondary {
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

  .mr-2 {
    margin-right: 0.5rem;
  }

  .mt-3 {
    margin-top: 1rem;
  }
}
</style>
