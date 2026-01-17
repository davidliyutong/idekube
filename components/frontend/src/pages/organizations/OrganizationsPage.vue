<template>
  <div class="organizations-page">
    <div class="page-header">
      <h1 class="page-title">组织管理</h1>
      <va-button @click="showCreateDialog = true">
        <va-icon name="add" class="mr-2" />
        新建组织
      </va-button>
    </div>

    <!-- Filters -->
    <va-card class="filters-card">
      <va-card-content>
        <div class="filters-row">
          <va-input
            v-model="searchQuery"
            placeholder="搜索组织名称..."
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

          <va-button @click="fetchOrganizations" preset="secondary">
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
            :items="filteredOrganizations"
            :columns="columns"
            striped
            hoverable
            :loading="loading"
          >
            <template #cell(name)="{ row }">
              <router-link :to="`/app/organizations/${row.id}`" class="link-cell">
                <strong>{{ row.name }}</strong>
              </router-link>
            </template>

            <template #cell(description)="{ row }">
              <span class="text-secondary">{{ row.description || '-' }}</span>
            </template>

            <template #cell(members)="{ row }">
              <span>{{ row.members?.length || 0 }} 人</span>
            </template>

            <template #cell(admins)="{ row }">
              <span>{{ row.admins?.length || 0 }} 人</span>
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
                  @click="editOrganization(row as any)"
                  :disabled="loading"
                >
                  <va-icon name="edit" />
                </va-button>
                <va-button
                  size="small"
                  preset="secondary"
                  @click="openMembersModal(row as any)"
                  :disabled="loading"
                >
                  <va-icon name="people" />
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
            共 {{ total }} 个组织
          </span>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create/Edit Modal -->
    <create-organization-modal
      v-model="showCreateDialog"
      :editing-organization="editingOrganization"
      @created="handleOrganizationCreated"
      @updated="handleOrganizationUpdated"
    />

    <!-- Members Modal -->
    <va-modal
      v-model="showMembersModal"
      title="管理组织成员"
      size="large"
      okText="关闭"
      @ok="showMembersModal = false"
    >
      <div class="members-content">
        <div v-if="membersModalOrganization" class="org-info">
          <p><strong>组织:</strong> {{ membersModalOrganization.name }}</p>
          <p><strong>成员数:</strong> {{ membersModalOrganization.members?.length || 0 }}</p>
        </div>

        <div class="members-list">
          <div v-if="membersModalOrganization?.members && membersModalOrganization.members.length > 0">
            <div
              v-for="member in membersModalOrganization.members"
              :key="member.id"
              class="member-item"
            >
              <div class="member-info">
                <strong>{{ member.username }}</strong>
                <span class="text-secondary">{{ member.email }}</span>
              </div>
              <div class="member-role">
                <va-chip
                  :color="member.role === 'admin' ? 'danger' : 'secondary'"
                  size="small"
                >
                  {{ getRoleLabel(member.role) }}
                </va-chip>
              </div>
              <div class="member-actions">
                <va-button
                  size="small"
                  preset="secondary"
                  @click="toggleMemberAdmin(member.id, member.role === 'admin')"
                >
                  <va-icon :name="member.role === 'admin' ? 'shield_off' : 'shield'" />
                </va-button>
                <va-button
                  size="small"
                  color="danger"
                  preset="secondary"
                  @click="removeMember(member.id)"
                >
                  <va-icon name="delete" />
                </va-button>
              </div>
            </div>
          </div>
          <div v-else class="empty-state">
            <p>暂无成员</p>
          </div>
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
      <p>确定要删除组织 <strong>{{ organizationToDelete?.name }}</strong> 吗？</p>
      <p class="text-secondary">此操作不可逆，请谨慎操作。</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useOrganizationStore, type Organization } from '@/stores/organization'
import CreateOrganizationModal from '@/components/CreateOrganizationModal.vue'
import { formatDate } from '@/utils/date'

const organizationStore = useOrganizationStore()

// Refs
const searchQuery = ref('')
const selectedStatus = ref<string | null>(null)
const showCreateDialog = ref(false)
const editingOrganization = ref<Organization | null>(null)
const showDeleteConfirm = ref(false)
const organizationToDelete = ref<Organization | null>(null)
const showMembersModal = ref(false)
const membersModalOrganization = ref<Organization | null>(null)

// Table columns
const columns: any[] = [
  { key: 'name', label: '组织名称', width: '200px' },
  { key: 'description', label: '描述', width: '250px' },
  { key: 'members', label: '成员数' },
  { key: 'admins', label: '管理员数' },
  { key: 'status', label: '状态' },
  { key: 'created_at', label: '创建时间' },
  { key: 'actions', label: '操作', width: '150px', align: 'center' },
]

const statusOptions = [
  { text: '活跃', value: 'active' },
  { text: '非活跃', value: 'inactive' },
  { text: '禁用', value: 'suspended' },
]

// Computed
const { organizations, loading, total, currentPage, pageSize } = organizationStore

const totalPages = computed(() => Math.ceil(total / pageSize) || 1)

const filteredOrganizations = computed(() => {
  let result = organizations

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter((o) => o.name.toLowerCase().includes(query))
  }

  if (selectedStatus.value) {
    result = result.filter((o) => o.status === selectedStatus.value)
  }

  return result
})

// Methods
function getRoleLabel(role: string): string {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    member: '普通成员',
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

async function fetchOrganizations() {
  await organizationStore.fetchOrganizations({
    page: currentPage,
    page_size: pageSize,
  })
}

function editOrganization(organization: Organization) {
  editingOrganization.value = organization
  showCreateDialog.value = true
}

function openMembersModal(organization: Organization) {
  membersModalOrganization.value = organization
  showMembersModal.value = true
}

async function toggleMemberAdmin(memberId: string, isCurrentlyAdmin: boolean) {
  if (membersModalOrganization.value) {
    await organizationStore.setMemberAdmin(
      membersModalOrganization.value.id,
      memberId,
      !isCurrentlyAdmin
    )
    // Refetch to update modal
    await organizationStore.fetchOrganization(membersModalOrganization.value.id)
    membersModalOrganization.value = organizationStore.currentOrganization
  }
}

async function removeMember(memberId: string) {
  if (membersModalOrganization.value) {
    await organizationStore.removeMember(membersModalOrganization.value.id, memberId)
    // Refetch to update modal
    await organizationStore.fetchOrganization(membersModalOrganization.value.id)
    membersModalOrganization.value = organizationStore.currentOrganization
  }
}

function handleDelete(organizationId: string) {
  organizationToDelete.value = organizations.find((o) => o.id === organizationId) || null
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (organizationToDelete.value) {
    await organizationStore.deleteOrganization(organizationToDelete.value.id)
    showDeleteConfirm.value = false
    organizationToDelete.value = null
    await fetchOrganizations()
  }
}

function handleOrganizationCreated() {
  editingOrganization.value = null
  showCreateDialog.value = false
  fetchOrganizations()
}

function handleOrganizationUpdated() {
  editingOrganization.value = null
  showCreateDialog.value = false
  fetchOrganizations()
}

function onPageChange(page: number) {
  organizationStore.currentPage = page
  fetchOrganizations()
}

// Lifecycle
onMounted(() => {
  fetchOrganizations()
})
</script>

<style lang="scss" scoped>
.organizations-page {
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

.members-content {
  padding: 1rem 0;

  .org-info {
    margin-bottom: 1.5rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--va-background-border);

    p {
      margin: 0.5rem 0;
      font-size: 0.875rem;
    }
  }

  .members-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;

    .member-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;

      .member-info {
        flex: 1;
        display: flex;
        flex-direction: column;
        gap: 0.25rem;

        strong {
          color: var(--va-textColor);
        }
      }

      .member-role {
        display: flex;
        justify-content: center;
      }

      .member-actions {
        display: flex;
        gap: 0.5rem;
      }
    }

    .empty-state {
      text-align: center;
      padding: 2rem 1rem;
      color: var(--va-textColorSecondary);
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
