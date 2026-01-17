<template>
  <div class="organization-detail-page">
    <div class="page-header">
      <router-link to="/app/organizations" class="back-link">
        <va-icon name="arrow_back" class="mr-2" />
        返回列表
      </router-link>
      <div class="header-actions">
        <va-button @click="showEditDialog = true" :disabled="loading">
          <va-icon name="edit" class="mr-2" />
          编辑
        </va-button>
        <va-button @click="showMembersModal = true" :disabled="loading">
          <va-icon name="people" class="mr-2" />
          管理成员
        </va-button>
        <va-button color="danger" @click="handleDelete" :disabled="loading">
          <va-icon name="delete" class="mr-2" />
          删除
        </va-button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <va-progress-circle indeterminate />
    </div>

    <!-- Content -->
    <div v-else-if="organization" class="org-content">
      <!-- Info Cards -->
      <div class="cards-grid">
        <!-- Basic Info Card -->
        <va-card class="info-card">
          <va-card-title>基本信息</va-card-title>
          <va-card-content>
            <div class="info-group">
              <label>组织名称</label>
              <p>{{ organization.name }}</p>
            </div>
            <div class="info-group">
              <label>描述</label>
              <p>{{ organization.description || '暂无描述' }}</p>
            </div>
            <div class="info-group">
              <label>状态</label>
              <p>
                <va-badge
                  :color="getStatusColor(organization.status)"
                  :text-color="organization.status === 'inactive' ? 'black' : 'white'"
                >
                  {{ getStatusLabel(organization.status) }}
                </va-badge>
              </p>
            </div>
            <div class="info-group">
              <label>创建时间</label>
              <p>{{ formatDate(organization.created_at) }}</p>
            </div>
          </va-card-content>
        </va-card>

        <!-- Quotas Card -->
        <va-card class="info-card">
          <va-card-title>配额信息</va-card-title>
          <va-card-content>
            <div class="info-group">
              <label>工作空间配额</label>
              <p>{{ organization.workspace_quota }} 个</p>
            </div>
            <div class="info-group">
              <label>存储配额</label>
              <p>{{ formatStorageQuota(organization.storage_quota || 0) }}</p>
            </div>
          </va-card-content>
        </va-card>

        <!-- Members Stats Card -->
        <va-card class="info-card">
          <va-card-title>成员统计</va-card-title>
          <va-card-content>
            <div class="info-group">
              <label>总成员数</label>
              <p class="stat-number">{{ organization.members?.length || 0 }}</p>
            </div>
            <div class="info-group">
              <label>管理员数</label>
              <p class="stat-number">{{ organization.admins?.length || 0 }}</p>
            </div>
          </va-card-content>
        </va-card>
      </div>

      <!-- Members Section -->
      <va-card class="members-card">
        <va-card-title>
          <div class="section-header">
            <span>成员管理</span>
            <va-button
              size="small"
              @click="showMembersModal = true"
            >
              <va-icon name="group_add" class="mr-2" />
              添加成员
            </va-button>
          </div>
        </va-card-title>
        <va-card-content>
          <div class="members-list">
            <div
              v-if="organization.members && organization.members.length > 0"
              class="members-table-wrapper"
            >
              <div class="members-header">
                <div class="member-col username">用户名</div>
                <div class="member-col email">邮箱</div>
                <div class="member-col role">角色</div>
                <div class="member-col actions">操作</div>
              </div>
              <div
                v-for="member in organization.members"
                :key="member.id"
                class="member-row"
              >
                <div class="member-col username">
                  <router-link :to="`/app/users/${member.id}`" class="member-link">
                    {{ member.username }}
                  </router-link>
                </div>
                <div class="member-col email">{{ member.email }}</div>
                <div class="member-col role">
                  <va-chip
                    :color="member.role === 'admin' ? 'danger' : 'secondary'"
                    size="small"
                  >
                    {{ getRoleLabel(member.role) }}
                  </va-chip>
                </div>
                <div class="member-col actions">
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
        </va-card-content>
      </va-card>
    </div>

    <!-- Not Found -->
    <div v-else class="not-found">
      <p>未找到该组织</p>
    </div>

    <!-- Edit Modal -->
    <create-organization-modal
      v-model="showEditDialog"
      :editing-organization="organization"
      @updated="handleOrganizationUpdated"
    />

    <!-- Members Management Modal -->
    <va-modal
      v-model="showMembersModal"
      title="管理组织成员"
      size="large"
      okText="关闭"
      @ok="showMembersModal = false"
    >
      <div class="members-modal-content">
        <div class="add-member-section">
          <div class="form-group">
            <label>选择成员</label>
            <va-select
              v-model="selectedMemberId"
              :options="availableUsers"
              text-by="username"
              track-by="id"
              placeholder="搜索用户..."
              searchable
              class="select-input"
            />
          </div>
          <va-button @click="addMember" :disabled="!selectedMemberId">
            <va-icon name="person_add" class="mr-2" />
            添加
          </va-button>
        </div>

        <div class="current-members">
          <h4>当前成员</h4>
          <div v-if="organization?.members && organization.members.length > 0">
            <div
              v-for="member in organization.members"
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
      <p>确定要删除组织 <strong>{{ organization?.name }}</strong> 吗？</p>
      <p class="text-secondary">此操作不可逆，请谨慎操作。</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useOrganizationStore } from '@/stores/organization'
import { useUserStore } from '@/stores/user'
import CreateOrganizationModal from '@/components/CreateOrganizationModal.vue'
import { formatDate } from '@/utils/date'

const router = useRouter()
const route = useRoute()
const organizationStore = useOrganizationStore()
const userStore = useUserStore()

// Refs
const showEditDialog = ref(false)
const showMembersModal = ref(false)
const showDeleteConfirm = ref(false)
const selectedMemberId = ref<string | null>(null)
const { loading } = organizationStore

// Computed
const organization = computed(() => organizationStore.currentOrganization)

const availableUsers = computed(() => {
  const memberIds = new Set(organization.value?.members?.map((m) => m.id) || [])
  return userStore.users.filter((u) => !memberIds.has(u.id))
})

// Methods
function formatStorageQuota(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`
}

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

async function fetchData() {
  const organizationId = route.params.id as string
  if (organizationId) {
    await organizationStore.fetchOrganization(organizationId)
    await userStore.fetchUsers({
      page: 1,
      page_size: 1000,
    })
  }
}

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (organization.value) {
    await organizationStore.deleteOrganization(organization.value.id)
    showDeleteConfirm.value = false
    await router.push('/app/organizations')
  }
}

function handleOrganizationUpdated() {
  showEditDialog.value = false
  fetchData()
}

async function addMember() {
  if (organization.value && selectedMemberId.value) {
    await organizationStore.addMember(organization.value.id, selectedMemberId.value)
    selectedMemberId.value = null
    await fetchData()
  }
}

async function removeMember(memberId: string) {
  if (organization.value) {
    await organizationStore.removeMember(organization.value.id, memberId)
    await fetchData()
  }
}

async function toggleMemberAdmin(memberId: string, isCurrentlyAdmin: boolean) {
  if (organization.value) {
    await organizationStore.setMemberAdmin(
      organization.value.id,
      memberId,
      !isCurrentlyAdmin
    )
    await fetchData()
  }
}

// Lifecycle
onMounted(() => {
  fetchData()
})
</script>

<style lang="scss" scoped>
.organization-detail-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--va-background-border);

    .back-link {
      color: var(--va-background-border);
      text-decoration: none;
      display: flex;
      align-items: center;
      gap: 0.5rem;

      &:hover {
        text-decoration: underline;
      }
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
    }
  }

  .loading-state {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 400px;
  }

  .org-content {
    .cards-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 1.5rem;
      margin-bottom: 2rem;

      .info-card {
        .info-group {
          margin-bottom: 1.5rem;

          &:last-child {
            margin-bottom: 0;
          }

          label {
            display: block;
            font-size: 0.875rem;
            color: var(--va-textColorSecondary);
            margin-bottom: 0.5rem;
            font-weight: 500;
          }

          p {
            margin: 0;
            font-size: 1rem;
            color: var(--va-textColor);

            &.stat-number {
              font-size: 2rem;
              font-weight: 600;
            }
          }
        }
      }
    }

    .members-card {
      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
      }

      .members-list {
        .members-table-wrapper {
          border: 1px solid var(--va-background-border);
          border-radius: 0.5rem;
          overflow: hidden;

          .members-header {
            display: grid;
            grid-template-columns: 1fr 1.5fr 1fr 150px;
            gap: 1rem;
            padding: 0.75rem 1rem;
            background-color: var(--va-background-shade);
            font-weight: 600;
            border-bottom: 1px solid var(--va-background-border);

            .member-col {
              font-size: 0.875rem;
            }
          }

          .member-row {
            display: grid;
            grid-template-columns: 1fr 1.5fr 1fr 150px;
            gap: 1rem;
            padding: 0.75rem 1rem;
            align-items: center;
            border-bottom: 1px solid var(--va-background-border);

            &:last-child {
              border-bottom: none;
            }

            .member-col {
              &.username {
                .member-link {
                  color: var(--va-background-border);
                  text-decoration: none;

                  &:hover {
                    text-decoration: underline;
                  }
                }
              }

              &.actions {
                display: flex;
                gap: 0.5rem;
              }
            }
          }
        }

        .empty-state {
          text-align: center;
          padding: 2rem;
          color: var(--va-textColorSecondary);
        }
      }
    }
  }

  .not-found {
    text-align: center;
    padding: 3rem 1rem;
    color: var(--va-textColorSecondary);
  }
}

.members-modal-content {
  padding: 1rem 0;

  .add-member-section {
    display: flex;
    gap: 1rem;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--va-background-border);

    .form-group {
      flex: 1;

      label {
        display: block;
        font-size: 0.875rem;
        font-weight: 500;
        margin-bottom: 0.5rem;
        color: var(--va-textColorSecondary);
      }

      .select-input {
        width: 100%;
      }
    }
  }

  .current-members {
    h4 {
      margin: 0 0 1rem 0;
      font-size: 1rem;
    }

    .member-item {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0.75rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;
      margin-bottom: 0.5rem;

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
      padding: 1.5rem;
      color: var(--va-textColorSecondary);
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;
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
