<template>
  <div class="workspace-transfers-page">
    <div class="page-header">
      <h1 class="page-title">Workspace Transfers</h1>
    </div>

    <!-- Pending Transfers Tab -->
    <va-card>
      <va-card-title>Pending Transfers</va-card-title>
      <va-card-content>
        <va-data-table
          :items="pendingTransfers"
          :columns="columns"
          :loading="loading"
          striped
          hoverable
        >
          <template #cell(workspace)="{ row }">
            <div class="workspace-info">
              <strong>{{ row.workspace_name }}</strong>
              <span class="workspace-id">ID: {{ row.workspace_id }}</span>
            </div>
          </template>

          <template #cell(from)="{ row }">
            <va-chip size="small" color="secondary">
              {{ row.from_user_name || row.from_user_id }}
            </va-chip>
          </template>

          <template #cell(to)="{ row }">
            <va-chip size="small" color="primary">
              {{ row.to_user_name || row.to_user_id }}
            </va-chip>
          </template>

          <template #cell(status)="{ row }">
            <va-badge
              :color="getStatusColor(row.status)"
              :text="row.status"
            />
          </template>

          <template #cell(requested_at)="{ row }">
            <span>{{ formatDate(row.requested_at) }}</span>
          </template>

          <template #cell(actions)="{ row }">
            <div class="action-buttons">
              <va-button
                v-if="row.status === 'pending' && isReceiver(row as any)"
                size="small"
                color="success"
                @click="handleRespond(row as any, 'accept')"
                :loading="respondingId === row.id"
              >
                <va-icon name="check" class="mr-1" />
                Accept
              </va-button>
              <va-button
                v-if="row.status === 'pending' && isReceiver(row as any)"
                size="small"
                color="danger"
                preset="secondary"
                @click="handleRespond(row as any, 'reject')"
                :loading="respondingId === row.id"
              >
                <va-icon name="close" class="mr-1" />
                Reject
              </va-button>
              <va-button
                v-if="row.status === 'pending' && isSender(row as any)"
                size="small"
                color="danger"
                preset="secondary"
                @click="handleCancel(row as any)"
                :loading="cancelingId === row.id"
              >
                <va-icon name="cancel" class="mr-1" />
                Cancel
              </va-button>
            </div>
          </template>
        </va-data-table>

        <div v-if="pendingTransfers.length === 0 && !loading" class="empty-state">
          <va-icon name="inbox" size="large" color="secondary" />
          <p>No pending transfers</p>
        </div>
      </va-card-content>
    </va-card>

    <!-- Initiate Transfer Card -->
    <va-card class="mt-4">
      <va-card-title>Initiate Workspace Transfer</va-card-title>
      <va-card-content>
        <div class="transfer-form">
          <va-select
            v-model="transferForm.workspace_id"
            label="Select Workspace"
            :options="ownedWorkspaces"
            text-by="name"
            value-by="id"
            placeholder="Choose a workspace to transfer"
          />

          <va-input
            v-model="transferForm.new_owner_id"
            label="New Owner User ID"
            placeholder="Enter the user ID of the new owner"
          />

          <va-button
            @click="initiateTransfer"
            :loading="initiating"
            :disabled="!transferForm.workspace_id || !transferForm.new_owner_id"
          >
            <va-icon name="send" class="mr-2" />
            Initiate Transfer
          </va-button>
        </div>
      </va-card-content>
    </va-card>

    <!-- Respond Confirmation Modal -->
    <va-modal
      v-model="showRespondModal"
      :title="respondAction === 'accept' ? 'Accept Transfer' : 'Reject Transfer'"
      okText="Confirm"
      @ok="confirmRespond"
    >
      <p v-if="respondAction === 'accept'">
        Are you sure you want to accept the transfer of workspace
        <strong>{{ transferToRespond?.workspace_name }}</strong>?
      </p>
      <p v-else>
        Are you sure you want to reject the transfer of workspace
        <strong>{{ transferToRespond?.workspace_name }}</strong>?
      </p>
      <p class="text-secondary mt-2">
        {{ respondAction === 'accept' ? 'You will become the new owner of this workspace.' : 'The transfer request will be declined.' }}
      </p>
    </va-modal>

    <!-- Cancel Confirmation Modal -->
    <va-modal
      v-model="showCancelModal"
      title="Cancel Transfer"
      okText="Cancel Transfer"
      cancelText="Close"
      @ok="confirmCancel"
    >
      <p>
        Are you sure you want to cancel the transfer of workspace
        <strong>{{ transferToCancel?.workspace_name }}</strong>?
      </p>
      <p class="text-secondary mt-2">This action cannot be undone.</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import apiClient from '@/api/client'
import { useWorkspaceStore } from '@/stores/workspace'
import { useAuthStore } from '@/stores/auth'
import { formatDate } from '@/utils/date'
import { useNotification } from '@/composables/useNotification'

interface WorkspaceTransfer {
  id: string
  workspace_id: string
  workspace_name: string
  from_user_id: string
  from_user_name?: string
  to_user_id: string
  to_user_name?: string
  status: 'pending' | 'accepted' | 'rejected' | 'cancelled'
  requested_at: string
  responded_at?: string
}

const workspaceStore = useWorkspaceStore()
const authStore = useAuthStore()
const { showSuccess, showError } = useNotification()

const loading = ref(false)
const initiating = ref(false)
const respondingId = ref<string | null>(null)
const cancelingId = ref<string | null>(null)
const pendingTransfers = ref<WorkspaceTransfer[]>([])

const showRespondModal = ref(false)
const showCancelModal = ref(false)
const transferToRespond = ref<WorkspaceTransfer | null>(null)
const transferToCancel = ref<WorkspaceTransfer | null>(null)
const respondAction = ref<'accept' | 'reject'>('accept')

const transferForm = ref({
  workspace_id: '',
  new_owner_id: '',
})

const columns = [
  { key: 'workspace', label: 'Workspace' },
  { key: 'from', label: 'From' },
  { key: 'to', label: 'To' },
  { key: 'status', label: 'Status' },
  { key: 'requested_at', label: 'Requested At' },
  { key: 'actions', label: 'Actions', width: '200px' },
]

const ownedWorkspaces = computed(() => {
  return workspaceStore.workspaces.filter(
    (w) => w.owner_id === authStore.user?.id
  )
})

function getStatusColor(status: string): string {
  const colors: Record<string, string> = {
    pending: 'warning',
    accepted: 'success',
    rejected: 'danger',
    cancelled: 'secondary',
  }
  return colors[status] || 'secondary'
}

function isReceiver(transfer: WorkspaceTransfer): boolean {
  return transfer.to_user_id === authStore.user?.id
}

function isSender(transfer: WorkspaceTransfer): boolean {
  return transfer.from_user_id === authStore.user?.id
}

async function fetchPendingTransfers() {
  loading.value = true
  try {
    const response = await apiClient.get<WorkspaceTransfer[]>('/workspace-transfers/pending')
    pendingTransfers.value = response.data
  } catch (err: any) {
    showError('Failed to fetch pending transfers')
    console.error('Fetch transfers error:', err)
  } finally {
    loading.value = false
  }
}

async function initiateTransfer() {
  initiating.value = true
  try {
    await apiClient.post(`/workspaces/${transferForm.value.workspace_id}/transfer`, {
      new_owner_id: transferForm.value.new_owner_id,
    })
    showSuccess('Transfer initiated successfully')
    transferForm.value = { workspace_id: '', new_owner_id: '' }
    await fetchPendingTransfers()
  } catch (err: any) {
    showError('Failed to initiate transfer')
    console.error('Initiate transfer error:', err)
  } finally {
    initiating.value = false
  }
}

function handleRespond(transfer: WorkspaceTransfer, action: 'accept' | 'reject') {
  transferToRespond.value = transfer
  respondAction.value = action
  showRespondModal.value = true
}

async function confirmRespond() {
  if (!transferToRespond.value) return

  respondingId.value = transferToRespond.value.id
  try {
    await apiClient.post(`/workspace-transfers/${transferToRespond.value.id}/respond`, {
      response: respondAction.value,
    })
    showSuccess(`Transfer ${respondAction.value}ed successfully`)
    await fetchPendingTransfers()
    showRespondModal.value = false
  } catch (err: any) {
    showError(`Failed to ${respondAction.value} transfer`)
    console.error('Respond to transfer error:', err)
  } finally {
    respondingId.value = null
    transferToRespond.value = null
  }
}

function handleCancel(transfer: WorkspaceTransfer) {
  transferToCancel.value = transfer
  showCancelModal.value = true
}

async function confirmCancel() {
  if (!transferToCancel.value) return

  cancelingId.value = transferToCancel.value.id
  try {
    await apiClient.post(`/workspace-transfers/${transferToCancel.value.id}/cancel`)
    showSuccess('Transfer cancelled successfully')
    await fetchPendingTransfers()
    showCancelModal.value = false
  } catch (err: any) {
    showError('Failed to cancel transfer')
    console.error('Cancel transfer error:', err)
  } finally {
    cancelingId.value = null
    transferToCancel.value = null
  }
}

onMounted(async () => {
  await workspaceStore.fetchWorkspaces()
  await fetchPendingTransfers()
})
</script>

<style lang="scss" scoped>
.workspace-transfers-page {
  .page-header {
    margin-bottom: 2rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .workspace-info {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;

    .workspace-id {
      font-size: 0.75rem;
      color: var(--va-textColorSecondary);
      font-family: 'Monaco', 'Menlo', monospace;
    }
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem 1rem;
    color: var(--va-textColorSecondary);

    p {
      margin-top: 1rem;
    }
  }

  .transfer-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    max-width: 600px;
  }
}

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-1 { margin-right: 0.25rem; }
.mr-2 { margin-right: 0.5rem; }
.mt-2 { margin-top: 0.5rem; }
.mt-4 { margin-top: 1rem; }
</style>
