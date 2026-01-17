<template>
  <div class="user-detail-page">
    <div class="page-header">
      <router-link to="/app/users" class="back-link">
        <va-icon name="arrow_back" class="mr-2" />
        返回列表
      </router-link>
      <h1 class="page-title">{{ user?.username || '加载中...' }}</h1>
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
          @click="openRolesModal()"
          preset="secondary"
          :disabled="loading || !user"
        >
          <va-icon name="security" class="mr-2" />
          管理角色
        </va-button>
        <va-button
          @click="handleDelete"
          color="danger"
          preset="secondary"
          :disabled="loading || !user"
        >
          <va-icon name="delete" class="mr-2" />
          删除
        </va-button>
      </div>
    </div>

    <va-progress v-if="loading" indeterminate class="mb-3" />

    <template v-if="user">
      <!-- Info Cards -->
      <div class="cards-grid">
        <!-- Basic Info -->
        <va-card class="info-card">
          <va-card-title>基本信息</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="label">用户ID</span>
              <span class="value">{{ user.id }}</span>
            </div>
            <div class="info-row">
              <span class="label">用户名</span>
              <span class="value">{{ user.username }}</span>
            </div>
            <div class="info-row">
              <span class="label">邮箱</span>
              <a :href="`mailto:${user.email}`" class="email-link">{{ user.email }}</a>
            </div>
            <div class="info-row">
              <span class="label">全名</span>
              <span class="value">{{ user.full_name || '-' }}</span>
            </div>
            <div class="info-row">
              <span class="label">状态</span>
              <va-badge
                :color="getStatusColor(user.status)"
                :text-color="getStatusColor(user.status) === 'warning' ? 'black' : 'white'"
              >
                {{ getStatusLabel(user.status) }}
              </va-badge>
            </div>
          </va-card-content>
        </va-card>

        <!-- Account Status -->
        <va-card class="info-card">
          <va-card-title>账户状态</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="label">管理员</span>
              <span class="value">
                <va-icon v-if="user.is_admin" name="verified" color="success" />
                <span v-else class="text-secondary">否</span>
              </span>
            </div>
            <div class="info-row">
              <span class="label">创建时间</span>
              <span class="value">{{ formatDate(user.created_at) }}</span>
            </div>
            <div class="info-row">
              <span class="label">更新时间</span>
              <span class="value">{{ formatDate(user.updated_at) }}</span>
            </div>
            <div v-if="user.last_login" class="info-row">
              <span class="label">最后登录</span>
              <span class="value">{{ formatDate(user.last_login) }}</span>
            </div>
          </va-card-content>
        </va-card>

        <!-- Statistics -->
        <va-card class="info-card">
          <va-card-title>统计信息</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="label">角色数量</span>
              <span class="value">{{ user.roles.length }}</span>
            </div>
            <div class="info-row">
              <span class="label">组织数量</span>
              <span class="value">{{ user.organization_ids?.length || 0 }}</span>
            </div>
          </va-card-content>
        </va-card>
      </div>

      <!-- Roles -->
      <va-card class="section-card">
        <va-card-title>分配的角色 ({{ user.roles.length }})</va-card-title>
        <va-card-content>
          <div class="roles-list">
            <div v-for="role in user.roles" :key="role" class="role-item">
              <va-icon name="check_circle" class="mr-2" color="success" />
              <span>{{ getRoleLabel(role) }}</span>
            </div>
          </div>
        </va-card-content>
      </va-card>

      <!-- Organizations -->
      <va-card v-if="user.organization_ids && user.organization_ids.length > 0" class="section-card">
        <va-card-title>所属组织 ({{ user.organization_ids.length }})</va-card-title>
        <va-card-content>
          <div class="organizations-list">
            <div
              v-for="orgId in user.organization_ids"
              :key="orgId"
              class="org-item"
            >
              <va-icon name="business" class="mr-2" />
              <router-link :to="`/app/organizations/${orgId}`" class="org-link">
                {{ orgId }}
              </router-link>
            </div>
          </div>
        </va-card-content>
      </va-card>
    </template>

    <!-- Edit Modal -->
    <create-user-modal
      v-model="showEditDialog"
      :editing-user="user"
      @updated="handleUserUpdated"
    />

    <!-- Roles Modal -->
    <va-modal
      v-model="showRolesModal"
      title="管理用户角色"
      okText="保存"
      cancelText="取消"
      @ok="saveRoles"
    >
      <div class="roles-content">
        <div v-if="user" class="user-info">
          <p><strong>用户:</strong> {{ user.username }}</p>
          <p><strong>邮箱:</strong> {{ user.email }}</p>
        </div>

        <div class="roles-grid">
          <va-checkbox
            v-for="role in availableRoles"
            :key="role"
            v-model="selectedRoles"
            :option="role"
            class="role-checkbox"
          >
            {{ getRoleLabel(role) }}
          </va-checkbox>
        </div>
      </div>
    </va-modal>

    <!-- Delete Confirmation Modal -->
    <va-modal
      v-model="showDeleteConfirm"
      title="确认删除"
      okText="删除"
      cancelText="取消"
      @ok="confirmDelete"
    >
      <p>确定要删除用户 <strong>{{ user?.username }}</strong> 吗？</p>
      <p class="text-secondary">此操作不可逆，请谨慎操作。</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import CreateUserModal from '@/components/CreateUserModal.vue'
import { formatDate } from '@/utils/date'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const id = computed(() => route.params.id as string)

// Refs
const showEditDialog = ref(false)
const showDeleteConfirm = ref(false)
const showRolesModal = ref(false)
const selectedRoles = ref<string[]>([])

// Computed
const user = computed(() => userStore.currentUser)
const loading = computed(() => userStore.loading)
const availableRoles = computed(() => userStore.availableRoles)

// Methods
function getRoleLabel(role: string): string {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    operator: '运维人员',
    developer: '开发者',
    viewer: '浏览者',
  }
  return roleMap[role] || role
}

function getStatusColor(status: string): string {
  const colorMap: Record<string, string> = {
    active: 'success',
    inactive: 'secondary',
    suspended: 'danger',
  }
  return colorMap[status] || 'secondary'
}

function getStatusLabel(status: string): string {
  const labelMap: Record<string, string> = {
    active: '活跃',
    inactive: '非活跃',
    suspended: '禁用',
  }
  return labelMap[status] || status
}

async function fetchUser() {
  try {
    await userStore.fetchUser(id.value)
  } catch (error) {
    console.error('Failed to fetch user:', error)
  }
}

function openRolesModal() {
  if (user.value) {
    selectedRoles.value = [...user.value.roles]
    showRolesModal.value = true
  }
}

async function saveRoles() {
  if (user.value) {
    await userStore.updateUserRoles(user.value.id, selectedRoles.value)
    showRolesModal.value = false
    fetchUser()
  }
}

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  try {
    await userStore.deleteUser(id.value)
    await router.push('/app/users')
  } catch (error) {
    console.error('Failed to delete user:', error)
  }
}

function handleUserUpdated() {
  showEditDialog.value = false
  fetchUser()
}

// Lifecycle
onMounted(() => {
  fetchUser()
})
</script>

<style lang="scss" scoped>
.user-detail-page {
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

        .email-link {
          color: var(--va-background-border);
          text-decoration: none;

          &:hover {
            text-decoration: underline;
          }
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

  .roles-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    .role-item {
      display: flex;
      align-items: center;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;
    }
  }

  .organizations-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    .org-item {
      display: flex;
      align-items: center;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;

      .org-link {
        color: var(--va-background-border);
        text-decoration: none;

        &:hover {
          text-decoration: underline;
        }
      }
    }
  }

  .roles-content {
    padding: 1rem 0;

    .user-info {
      margin-bottom: 1.5rem;
      padding-bottom: 1rem;
      border-bottom: 1px solid var(--va-background-border);

      p {
        margin: 0.5rem 0;
        font-size: 0.875rem;
      }
    }

    .roles-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 1rem;

      .role-checkbox {
        padding: 0.5rem;
      }
    }
  }
}

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-2 {
  margin-right: 0.5rem;
}
</style>
