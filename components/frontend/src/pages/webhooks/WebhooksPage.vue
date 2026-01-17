<template>
  <div class="webhooks-page">
    <div class="page-header">
      <h1 class="page-title">Webhooks</h1>
      <va-button @click="openCreateModal()">
        <va-icon name="add" class="mr-2" />
        Create Webhook
      </va-button>
    </div>

    <!-- Webhooks Table -->
    <va-card>
      <va-card-content>
        <va-data-table
          :items="webhooks"
          :columns="columns"
          :loading="loading"
          striped
          hoverable
        >
          <template #cell(url)="{ row }">
            <code class="url-cell">{{ row.url }}</code>
          </template>

          <template #cell(events)="{ row }">
            <div class="events-cell">
              <va-chip
                v-for="event in row.events.slice(0, 2)"
                :key="event"
                size="small"
                class="event-chip"
              >
                {{ event }}
              </va-chip>
              <va-chip v-if="row.events.length > 2" size="small" color="secondary">
                +{{ row.events.length - 2 }} more
              </va-chip>
            </div>
          </template>

          <template #cell(enabled)="{ row }">
            <va-badge
              :color="row.enabled ? 'success' : 'danger'"
              :text="row.enabled ? 'Enabled' : 'Disabled'"
            />
          </template>

          <template #cell(last_triggered_at)="{ row }">
            <span>{{ row.last_triggered_at ? formatDate(row.last_triggered_at) : 'Never' }}</span>
          </template>

          <template #cell(actions)="{ row }">
            <div class="action-buttons">
              <va-button
                size="small"
                preset="secondary"
                @click="handleTest(row as any)"
                :loading="testingId === row.id"
              >
                <va-icon name="bug_report" />
              </va-button>
              <va-button
                size="small"
                preset="secondary"
                @click="openEditModal(row as any)"
              >
                <va-icon name="edit" />
              </va-button>
              <va-button
                size="small"
                color="danger"
                preset="secondary"
                @click="handleDelete(row as any)"
                :disabled="loading"
              >
                <va-icon name="delete" />
              </va-button>
            </div>
          </template>
        </va-data-table>
      </va-card-content>
    </va-card>

    <!-- Create/Edit Webhook Modal -->
    <va-modal
      v-model="showModal"
      :title="isEdit ? 'Edit Webhook' : 'Create Webhook'"
      size="large"
      okText="Save"
      @ok="handleSave"
    >
      <div class="webhook-form">
        <va-input
          v-model="formData.url"
          label="Webhook URL"
          placeholder="https://example.com/webhook"
        />

        <div class="form-group">
          <label class="va-input-label">Events</label>
          <div class="events-selection">
            <va-checkbox
              v-for="event in availableEvents"
              :key="event"
              v-model="formData.events"
              :array-value="event"
              class="event-checkbox"
            >
              {{ event }}
            </va-checkbox>
          </div>
        </div>

        <va-checkbox v-model="formData.enabled">
          Enabled
        </va-checkbox>
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
      <p>Are you sure you want to delete this webhook?</p>
      <p class="url-text"><strong>{{ webhookToDelete?.url }}</strong></p>
      <p class="text-secondary">This action cannot be undone.</p>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useWebhookStore, type Webhook } from '@/stores/webhook'
import { formatDate } from '@/utils/date'

const webhookStore = useWebhookStore()

const showModal = ref(false)
const showDeleteConfirm = ref(false)
const isEdit = ref(false)
const editingId = ref<number | null>(null)
const webhookToDelete = ref<Webhook | null>(null)
const testingId = ref<number | null>(null)

const formData = ref({
  url: '',
  events: [] as string[],
  enabled: true,
})

const availableEvents = [
  'workspace.created', 'workspace.started', 'workspace.stopped', 'workspace.deleted',
  'template.created', 'template.updated', 'template.deleted',
  'user.created', 'user.updated', 'user.deleted',
  'organization.created', 'organization.updated', 'organization.deleted',
]

const { webhooks, loading } = webhookStore

const columns = [
  { key: 'url', label: 'URL' },
  { key: 'events', label: 'Events' },
  { key: 'enabled', label: 'Status' },
  { key: 'last_triggered_at', label: 'Last Triggered' },
  { key: 'actions', label: 'Actions', width: '180px' },
]

function openCreateModal() {
  isEdit.value = false
  editingId.value = null
  formData.value = { url: '', events: [], enabled: true }
  showModal.value = true
}

function openEditModal(webhook: Webhook) {
  isEdit.value = true
  editingId.value = webhook.id
  formData.value = {
    url: webhook.url,
    events: [...webhook.events],
    enabled: webhook.enabled,
  }
  showModal.value = true
}

async function handleSave() {
  try {
    if (isEdit.value && editingId.value) {
      await webhookStore.updateWebhook(editingId.value, formData.value)
    } else {
      await webhookStore.createWebhook(formData.value)
    }
    showModal.value = false
  } catch (error) {
    console.error('Failed to save webhook:', error)
  }
}

function handleDelete(webhook: Webhook) {
  webhookToDelete.value = webhook
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  if (webhookToDelete.value) {
    await webhookStore.deleteWebhook(webhookToDelete.value.id)
    showDeleteConfirm.value = false
    webhookToDelete.value = null
  }
}

async function handleTest(webhook: Webhook) {
  testingId.value = webhook.id
  try {
    await webhookStore.testWebhook(webhook.id)
  } finally {
    testingId.value = null
  }
}

onMounted(() => {
  webhookStore.fetchWebhooks()
})
</script>

<style lang="scss" scoped>
.webhooks-page {
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

  .url-cell {
    font-family: 'Monaco', 'Menlo', monospace;
    font-size: 0.875rem;
    background-color: var(--va-background-shade);
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
  }

  .events-cell {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .action-buttons {
    display: flex;
    gap: 0.5rem;
  }
}

.webhook-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;

  .form-group {
    .va-input-label {
      display: block;
      margin-bottom: 0.5rem;
      font-weight: 500;
    }

    .events-selection {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
      gap: 0.5rem;
      padding: 1rem;
      background-color: var(--va-background-shade);
      border-radius: 0.5rem;

      .event-checkbox {
        margin: 0;
      }
    }
  }
}

.url-text {
  word-break: break-all;
  font-family: 'Monaco', 'Menlo', monospace;
  font-size: 0.875rem;
}

.text-secondary {
  color: var(--va-textColorSecondary);
}

.mr-2 { margin-right: 0.5rem; }
</style>
