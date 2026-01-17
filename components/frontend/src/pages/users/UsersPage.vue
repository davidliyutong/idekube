<template>
  <div class="users-page">
    <div class="page-header">
      <h1 class="page-title">用户管理</h1>
      <va-button @click="showCreateDialog = true">
        <va-icon name="add" class="mr-2" />
        新建用户
      </va-button>
    </div>

    <!-- Filters -->
    <va-card class="filters-card">
      <va-card-content>
        <div class="filters-row">
          <va-input
            v-model="searchQuery"
            placeholder="搜索用户名或邮箱..."
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

          <va-select
            v-model="selectedRole"
            :options="roleOptions"
            label="角色筛选"
            class="filter-item"
            clearable
          />

          <va-button @click="fetchUsers" preset="secondary">
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
            :items="filteredUsers"
            :columns="columns"
            striped
            hoverable
            :loading="loading"
          >
            <template #cell(username)="{ row }">
              <router-link :to="`/app/users/${row.id}`" class="link-cell">
                <strong>{{ row.username }}</strong>
              </router-link>
            </template>

            <template #cell(email)="{ row }">
              <a :href="`mailto:${row.email}`" class="email-link">{{ row.email }}</a>
            </template>

            <template #cell(full_name)="{ row }">
              <span>{{ row.full_name || '-' }}</span>
            </template>

            <template #cell(roles)="{ row }">
              <div class="roles-list">
                <va-chip
                  v-for="role in row.roles"
                  :key="role"
                  size="small"
                  class="role-chip"
                >
                  {{ getRoleLabel(role) }}
                </va-chip>
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

            <template #cell(is_admin)="{ row }">
              <va-icon v-if="row.is_admin" name="verified" color="success" />
              <span v-else class="text-secondary">-</span>
            </template>

            <template #cell(created_at)="{ row }">
              <span>{{ formatDate(row.created_at) }}</span>
            </template>

            <template #cell(actions)="{ row }">
              <div class="action-buttons">
                <va-button
                  size="small"
                  preset="secondary"
                  @click="editUser(row as any)"
                  :disabled="loading"
                >
                  <va-icon name="edit" />
                </va-button>
                <va-button
                  size="small"
                  preset="secondary"
                  @click="openRolesModal(row as any)"
                  :disabled="loading"
                >
                  <va-icon name="security" />
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
            共 {{ total }} 个用户
          </span>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create/Edit Modal -->
    <create-user-modal
      v-model="showCreateDialog"
      :editing-user="editingUser"
      @created="handleUserCreated"
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
        <div v-if="rolesModalUser" class="user-info">
          <p><strong>用户:</strong> {{ rolesModalUser.username }}</p>
          <p><strong>邮箱:</strong> {{ rolesModalUser.email }}</p>
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
      <p>确定要删除用户 <strong>{{ userToDelete?.username }}</strong> 吗？</p>
      <p class="text-secondary">此操作不可逆，请谨慎操作。</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore, type User } from '@/stores/user'
import CreateUserModal from '@/components/CreateUserModal.vue'
import { formatDate } from '@/utils/date'

const userStore = useUserStore()

// Refs
const searchQuery = ref('')
const selectedStatus = ref<string | null>(null)
const selectedRole = ref<string | null>(null)
const showCreateDialog = ref(false)
const editingUser = ref<User | null>(null)
const showDeleteConfirm = ref(false)
const userToDelete = ref<User | null>(null)
const showRolesModal = ref(false)
const rolesModalUser = ref<User | null>(null)
const selectedRoles = ref<string[]>([])

// Table columns
const columns: any[] = [
  { key: 'username', label: '用户名', width: '150px' },
  { key: 'email', label: '邮箱', width: '200px' },
  { key: 'full_name', label: '全名', width: '150px' },
  { key: 'roles', label: '角色', width: '200px' },
  { key: 'status', label: '状态' },
  { key: 'is_admin', label: '管理员', width: '80px' },
  { key: 'created_at', label: '创建时间' },
  { key: 'actions', label: '操作', width: '150px', align: 'center' },
]

const statusOptions = [
  { text: '活跃', value: 'active' },
  { text: '非活跃', value: 'inactive' },
  { text: '禁用', value: 'suspended' },
]

const roleOptions = computed(() =>
  availableRoles.map((role: string) => ({
    text: getRoleLabel(role),
    value: role,
  }))
)

// Computed
const { users, loading, total, currentPage, pageSize, availableRoles } = userStore

const totalPages = computed(() => Math.ceil(total / pageSize) || 1)

const filteredUsers = computed(() => {
  let result = users

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(
      (u) => u.username.toLowerCase().includes(query) || u.email.toLowerCase().includes(query)
    )
  }

  if (selectedStatus.value) {
    result = result.filter((u) => u.status === selectedStatus.value)
  }

  if (selectedRole.value) {
    result = result.filter((u) => u.roles.includes(selectedRole.value!))
  }

  return result
})

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

async function fetchUsers() {
  await userStore.fetchUsers({
    page: currentPage,
    page_size: pageSize,
  })
}

function editUser(user: User) {
  editingUser.value = user
  showCreateDialog.value = true
}

function openRolesModal(user: User) {
  rolesModalUser.value = user
  selectedRoles.value = [...user.roles]
  showRolesModal.value = true
}

async function saveRoles() {
  if (rolesModalUser.value) {
    await userStore.updateUserRoles(rolesModalUser.value.id, selectedRoles.value)
    showRolesModal.value = false
  }
}

function handleDelete(userId: string) {
  userToDelete.value = users.find((u) => u.id === userId) || null
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (userToDelete.value) {
    await userStore.deleteUser(userToDelete.value.id)
    showDeleteConfirm.value = false
    userToDelete.value = null
    await fetchUsers()
  }
}

function handleUserCreated() {
  editingUser.value = null
  showCreateDialog.value = false
  fetchUsers()
}

function handleUserUpdated() {
  editingUser.value = null
  showCreateDialog.value = false
  fetchUsers()
}

function onPageChange(page: number) {
  userStore.currentPage = page
  fetchUsers()
}

// Lifecycle
onMounted(() => {
  fetchUsers()
})
</script>

<style lang="scss" scoped>
.users-page {
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

    .email-link {
      color: var(--va-background-border);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }

    .roles-list {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;

      .role-chip {
        background-color: var(--va-background-shade);
      }
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

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-2 {
  margin-right: 0.5rem;
}
</style>
