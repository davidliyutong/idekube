<template>
  <div class="dashboard-page">
    <!-- Header -->
    <div class="dashboard-header">
      <div class="header-content">
        <h1 class="page-title">仪表盘</h1>
        <p class="subtitle">欢迎回来，{{ user?.username }}! 👋</p>
      </div>
      <div class="header-meta">
        <span class="date-time">{{ currentDateTime }}</span>
      </div>
    </div>

    <!-- Statistics Cards -->
    <div class="stats-grid">
      <div class="stat-card workspace-card" @click="navigateTo('workspaces')">
        <div class="card-header">
          <va-icon name="computer" class="card-icon" />
          <span class="card-title">工作空间</span>
        </div>
        <div class="card-value">{{ workspaceCount }}</div>
        <div class="card-subtitle">
          <span v-if="activeWorkspaces > 0" class="highlight">
            {{ activeWorkspaces }} 活跃
          </span>
          <span v-else class="text-secondary">无活跃工作空间</span>
        </div>
        <div class="card-footer">
          <va-button preset="plain" size="small" @click.stop="navigateTo('workspaces')">
            查看全部 →
          </va-button>
        </div>
      </div>

      <div class="stat-card template-card" @click="navigateTo('templates')">
        <div class="card-header">
          <va-icon name="layers" class="card-icon" />
          <span class="card-title">模板库</span>
        </div>
        <div class="card-value">{{ templateCount }}</div>
        <div class="card-subtitle">
          <span class="text-secondary">{{ publicTemplates }} 公开模板</span>
        </div>
        <div class="card-footer">
          <va-button preset="plain" size="small" @click.stop="navigateTo('templates')">
            浏览模板 →
          </va-button>
        </div>
      </div>

      <div class="stat-card volume-card" @click="navigateTo('volumes')">
        <div class="card-header">
          <va-icon name="storage" class="card-icon" />
          <span class="card-title">数据卷</span>
        </div>
        <div class="card-value">{{ volumeCount }}</div>
        <div class="card-subtitle">
          <span class="text-secondary">{{ formatSize(totalVolumeSize) }}</span>
        </div>
        <div class="card-footer">
          <va-button preset="plain" size="small" @click.stop="navigateTo('volumes')">
            管理卷 →
          </va-button>
        </div>
      </div>

      <div class="stat-card storage-card">
        <div class="card-header">
          <va-icon name="pie_chart" class="card-icon" />
          <span class="card-title">存储使用</span>
        </div>
        <div class="card-value">{{ storageUsagePercent }}%</div>
        <va-progress :model-value="storageUsagePercent" class="usage-progress" />
        <div class="card-subtitle">
          <span class="text-secondary">{{ formatSize(usedStorage) }} / {{ formatSize(totalStorage) }}</span>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <va-card class="quick-actions-card">
      <va-card-title>快速操作</va-card-title>
      <va-card-content>
        <div class="action-buttons">
          <va-button @click="createWorkspace" preset="primary">
            <va-icon name="add" class="mr-2" />
            新建工作空间
          </va-button>
          <va-button @click="navigateTo('templates')" preset="secondary">
            <va-icon name="layers" class="mr-2" />
            浏览模板库
          </va-button>
          <va-button @click="navigateTo('volumes')" preset="secondary">
            <va-icon name="storage" class="mr-2" />
            管理数据卷
          </va-button>
          <va-button @click="navigateTo('settings')" preset="secondary">
            <va-icon name="settings" class="mr-2" />
            设置
          </va-button>
        </div>
      </va-card-content>
    </va-card>

    <!-- Recent Activity -->
    <va-card class="activity-card">
      <va-card-title>最近活动</va-card-title>
      <va-card-content>
        <div v-if="recentActivities.length > 0" class="activity-list">
          <div v-for="(activity, index) in recentActivities" :key="index" class="activity-item">
            <div class="activity-icon" :class="activity.type">
              <va-icon :name="getActivityIcon(activity.type)" />
            </div>
            <div class="activity-content">
              <div class="activity-title">{{ activity.title }}</div>
              <div class="activity-description">{{ activity.description }}</div>
            </div>
            <div class="activity-time">{{ formatRelativeTime(activity.timestamp) }}</div>
          </div>
        </div>
        <div v-else class="empty-state">
          <va-icon name="history" class="empty-icon" />
          <p>暂无活动记录</p>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create Workspace Modal -->
    <create-workspace-modal
      v-model="showCreateWorkspaceModal"
      @created="handleWorkspaceCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWorkspaceStore } from '@/stores/workspace'
import { useTemplateStore } from '@/stores/template'
import { useVolumeStore } from '@/stores/volume'
import CreateWorkspaceModal from '@/components/CreateWorkspaceModal.vue'
import { formatDate } from '@/utils/date'

const router = useRouter()
const authStore = useAuthStore()
const workspaceStore = useWorkspaceStore()
const templateStore = useTemplateStore()
const volumeStore = useVolumeStore()

// Refs
const showCreateWorkspaceModal = ref(false)
const currentDateTime = ref('')

// Computed properties for statistics
const user = computed(() => authStore.user)

const workspaceCount = computed(() => workspaceStore.total || 0)

const activeWorkspaces = computed(() => {
  return workspaceStore.workspaces.filter((w: any) => w.status === 'running').length
})

const templateCount = computed(() => templateStore.total || 0)

const publicTemplates = computed(() => {
  return templateStore.templates.filter((t: any) => t.visibility === 'public').length
})

const volumeCount = computed(() => volumeStore.total || 0)

const totalVolumeSize = computed(() => {
  return volumeStore.volumes.reduce((sum: number, v: any) => sum + (v.size || 0), 0)
})

const usedStorage = computed(() => {
  return totalVolumeSize.value
})

const totalStorage = computed(() => {
  return 1024 * 1024 * 1024 * 1024 // 1TB default
})

const storageUsagePercent = computed(() => {
  const percent = Math.round((usedStorage.value / totalStorage.value) * 100)
  return Math.min(percent, 100)
})

// Recent activities
const recentActivities = computed(() => [
  {
    type: 'workspace',
    title: '工作空间创建',
    description: `您有 ${workspaceCount.value} 个工作空间`,
    timestamp: new Date().toISOString(),
  },
  {
    type: 'template',
    title: '模板库',
    description: `您可以使用 ${templateCount.value} 个模板快速创建工作空间`,
    timestamp: new Date().toISOString(),
  },
  {
    type: 'volume',
    title: '数据存储',
    description: `已使用 ${formatSize(usedStorage.value)} 的存储空间`,
    timestamp: new Date().toISOString(),
  },
])

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

function formatRelativeTime(timestamp: string): string {
  try {
    const now = new Date()
    const then = new Date(timestamp)
    const diff = now.getTime() - then.getTime()

    const minutes = Math.floor(diff / 60000)
    const hours = Math.floor(diff / 3600000)
    const days = Math.floor(diff / 86400000)

    if (minutes < 1) return '刚刚'
    if (minutes < 60) return `${minutes} 分钟前`
    if (hours < 24) return `${hours} 小时前`
    if (days < 7) return `${days} 天前`

    return formatDate(timestamp, 'MM/dd')
  } catch {
    return '未知'
  }
}

function getActivityIcon(type: string): string {
  const iconMap: Record<string, string> = {
    workspace: 'computer',
    template: 'layers',
    volume: 'storage',
  }
  return iconMap[type] || 'info'
}

function navigateTo(route: string) {
  router.push({ name: route })
}

function createWorkspace() {
  showCreateWorkspaceModal.value = true
}

function handleWorkspaceCreated() {
  showCreateWorkspaceModal.value = false
  workspaceStore.fetchWorkspaces()
}

function updateDateTime() {
  const now = new Date()
  const options: Intl.DateTimeFormatOptions = {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }
  currentDateTime.value = now.toLocaleDateString('zh-CN', options)
}

// Lifecycle
onMounted(async () => {
  updateDateTime()
  setInterval(updateDateTime, 60000) // Update every minute

  // Fetch initial data
  await Promise.all([
    workspaceStore.fetchWorkspaces({ page_size: 100 }),
    templateStore.fetchTemplates({ page_size: 100 }),
    volumeStore.fetchVolumes({ page_size: 100 }),
  ])
})
</script>

<style lang="scss" scoped>
.dashboard-page {
  padding: 0;

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 2rem;

    .header-content {
      flex: 1;

      .page-title {
        margin: 0 0 0.5rem 0;
        font-size: 2.5rem;
        font-weight: 700;
        background: linear-gradient(135deg, var(--va-background-border) 0%, var(--va-primary) 100%);
        background-clip: text;
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
      }

      .subtitle {
        margin: 0;
        color: var(--va-textColorSecondary);
        font-size: 1rem;
      }
    }

    .header-meta {
      text-align: right;

      .date-time {
        color: var(--va-textColorSecondary);
        font-size: 0.875rem;
      }
    }
  }

  // Statistics Grid
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;

    .stat-card {
      padding: 1.5rem;
      border-radius: 0.75rem;
      background: var(--va-background-element);
      border: 1px solid var(--va-background-border);
      cursor: pointer;
      transition: all 0.3s ease;
      display: flex;
      flex-direction: column;
      gap: 1rem;

      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
      }

      .card-header {
        display: flex;
        align-items: center;
        gap: 0.75rem;

        .card-icon {
          font-size: 1.5rem;
        }

        .card-title {
          font-size: 0.875rem;
          font-weight: 600;
          color: var(--va-textColorSecondary);
        }
      }

      .card-value {
        font-size: 2.5rem;
        font-weight: 700;
        color: var(--va-textColor);
        line-height: 1;
      }

      .card-subtitle {
        font-size: 0.875rem;
        color: var(--va-textColorSecondary);

        .highlight {
          color: var(--va-success);
          font-weight: 600;
        }

        .text-secondary {
          color: var(--va-textColorSecondary);
        }
      }

      .usage-progress {
        margin: 0.5rem 0;
      }

      .card-footer {
        margin-top: auto;
      }

      // Card color variants
      &.workspace-card {
        border-left: 4px solid var(--va-primary);

        .card-icon {
          color: var(--va-primary);
        }
      }

      &.template-card {
        border-left: 4px solid var(--va-success);

        .card-icon {
          color: var(--va-success);
        }
      }

      &.volume-card {
        border-left: 4px solid var(--va-info);

        .card-icon {
          color: var(--va-info);
        }
      }

      &.storage-card {
        border-left: 4px solid var(--va-warning);

        .card-icon {
          color: var(--va-warning);
        }
      }
    }
  }

  // Quick Actions Card
  .quick-actions-card {
    margin-bottom: 2rem;

    .va-card__title {
      margin-bottom: 1rem;
    }

    .action-buttons {
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;

      button {
        flex: 1;
        min-width: 200px;
      }
    }
  }

  // Activity Card
  .activity-card {
    .va-card__title {
      margin-bottom: 1.5rem;
    }

    .activity-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;

      .activity-item {
        display: flex;
        align-items: flex-start;
        gap: 1rem;
        padding: 1rem;
        background-color: var(--va-background-shade);
        border-radius: 0.5rem;
        transition: all 0.3s ease;

        &:hover {
          background-color: var(--va-background-border);
        }

        .activity-icon {
          display: flex;
          align-items: center;
          justify-content: center;
          width: 2.5rem;
          height: 2.5rem;
          border-radius: 50%;
          flex-shrink: 0;
          font-size: 1.25rem;

          &.workspace {
            background-color: rgba(var(--va-primary-rgb), 0.1);
            color: var(--va-primary);
          }

          &.template {
            background-color: rgba(var(--va-success-rgb), 0.1);
            color: var(--va-success);
          }

          &.volume {
            background-color: rgba(var(--va-info-rgb), 0.1);
            color: var(--va-info);
          }
        }

        .activity-content {
          flex: 1;
          min-width: 0;

          .activity-title {
            font-weight: 600;
            color: var(--va-textColor);
            margin-bottom: 0.25rem;
          }

          .activity-description {
            font-size: 0.875rem;
            color: var(--va-textColorSecondary);
          }
        }

        .activity-time {
          font-size: 0.75rem;
          color: var(--va-textColorSecondary);
          flex-shrink: 0;
        }
      }
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 3rem 1rem;
      color: var(--va-textColorSecondary);

      .empty-icon {
        font-size: 3rem;
        margin-bottom: 1rem;
        opacity: 0.5;
      }

      p {
        margin: 0;
        font-size: 0.875rem;
      }
    }
  }
}

// Utility classes
.mr-2 {
  margin-right: 0.5rem;
}

.text-secondary {
  color: var(--va-textColorSecondary);
}
</style>
