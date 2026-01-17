<template>
  <div class="policies-page">
    <div class="page-header">
      <h1 class="page-title">Policy Management</h1>
      <va-button @click="showCreateModal = true">
        <va-icon name="add" class="mr-2" />
        Add Policy
      </va-button>
    </div>

    <!-- Stats Card -->
    <va-card class="mb-4">
      <va-card-content>
        <div class="stat-content">
          <va-icon name="policy" size="large" color="primary" />
          <div class="stat-info">
            <h3>{{ policyCount }}</h3>
            <p>Total Policies</p>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Policies Table -->
    <va-card>
      <va-card-content>
        <va-data-table
          :items="policies"
          :columns="columns"
          :loading="loading"
          striped
          hoverable
        >
          <template #cell(subject)="{ row }">
            <va-chip size="small" color="primary">
              {{ row.subject }}
            </va-chip>
          </template>

          <template #cell(object)="{ row }">
            <code class="resource-code">{{ row.object }}</code>
          </template>

          <template #cell(action)="{ row }">
            <va-chip size="small" color="success">
              {{ row.action }}
            </va-chip>
          </template>

          <template #cell(actions)="{ row }">
            <va-button
              size="small"
              color="danger"
              preset="secondary"
              @click="handleDelete(row as any)"
              :disabled="loading"
            >
              <va-icon name="delete" />
            </va-button>
          </template>
        </va-data-table>
      </va-card-content>
    </va-card>

    <!-- Create Policy Modal -->
    <va-modal
      v-model="showCreateModal"
      title="Add Policy"
      size="medium"
      okText="Add"
      @ok="handleCreate"
    >
      <div class="policy-form">
        <va-input
          v-model="formData.subject"
          label="Subject"
          placeholder="user:123 or role:admin"
        />

        <va-input
          v-model="formData.object"
          label="Object (Resource)"
          placeholder="workspace:456 or template:*"
        />

        <va-select
          v-model="formData.action"
          label="Action"
          :options="availableActions"
          placeholder="Select an action"
        />

        <va-alert color="info" class="mt-2">
          <template #title>Policy Format</template>
          <ul class="policy-help">
            <li><strong>Subject:</strong> user:ID, role:NAME, or org:ID</li>
            <li><strong>Object:</strong> workspace:ID, template:ID, or use * for all</li>
            <li><strong>Action:</strong> read, write, delete, start, stop, etc.</li>
          </ul>
        </va-alert>
      </div>
    </va-modal>

    <!-- Delete Confirmation -->
    <va-modal
      v-model="showDeleteConfirm"
      title="Confirm Delete"
      okText="Delete"
      cancelText="Cancel"
      @ok="confirmDelete"
    >
      <p>Are you sure you want to delete this policy?</p>
      <div v-if="policyToDelete" class="policy-preview">
        <div class="preview-item">
          <strong>Subject:</strong> {{ policyToDelete.subject }}
        </div>
        <div class="preview-item">
          <strong>Object:</strong> {{ policyToDelete.object }}
        </div>
        <div class="preview-item">
          <strong>Action:</strong> {{ policyToDelete.action }}
        </div>
      </div>
      <p class="text-secondary mt-2">This action cannot be undone.</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usePolicyStore, type Policy } from '@/stores/policy'

const policyStore = usePolicyStore()

const showCreateModal = ref(false)
const showDeleteConfirm = ref(false)
const policyToDelete = ref<Policy | null>(null)

const formData = ref({
  subject: '',
  object: '',
  action: '',
})

const availableActions = [
  'read', 'write', 'delete',
  'start', 'stop', 'restart',
  'create', 'update',
  'manage', '*',
]

const { policies, loading, policyCount } = policyStore

const columns = [
  { key: 'subject', label: 'Subject' },
  { key: 'object', label: 'Object' },
  { key: 'action', label: 'Action' },
  { key: 'actions', label: 'Actions', width: '100px' },
]

async function handleCreate() {
  try {
    await policyStore.createPolicy(formData.value)
    showCreateModal.value = false
    formData.value = { subject: '', object: '', action: '' }
  } catch (error) {
    console.error('Failed to create policy:', error)
  }
}

function handleDelete(policy: Policy) {
  policyToDelete.value = policy
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (policyToDelete.value) {
    await policyStore.deletePolicy(
      policyToDelete.value.subject,
      policyToDelete.value.object,
      policyToDelete.value.action
    )
    showDeleteConfirm.value = false
    policyToDelete.value = null
  }
}

onMounted(() => {
  policyStore.fetchPolicies()
})
</script>

<style lang="scss" scoped>
.policies-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .stat-content {
    display: flex;
    align-items: center;
    gap: 1rem;

    .stat-info h3 {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }

    .stat-info p {
      margin: 0;
      color: var(--va-textColorSecondary);
      font-size: 0.875rem;
    }
  }

  .resource-code {
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 0.875rem;
    background-color: var(--va-background-shade);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
  }
}

.policy-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;

  .policy-help {
    margin: 0.5rem 0;
    padding-left: 1.25rem;
    font-size: 0.875rem;

    li {
      margin: 0.25rem 0;
    }
  }
}

.policy-preview {
  padding: 1rem;
  background-color: var(--va-background-shade);
  border-radius: 0.5rem;
  margin: 1rem 0;

  .preview-item {
    margin: 0.5rem 0;
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 0.875rem;

    strong {
      display: inline-block;
      min-width: 80px;
    }
  }
}

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-2 { margin-right: 0.5rem; }
.mb-4 { margin-bottom: 1rem; }
.mt-2 { margin-top: 0.5rem; }
</style>
