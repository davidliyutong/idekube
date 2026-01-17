<template>
  <div class="template-detail-page">
    <div v-if="templateStore.loading && !template" class="loading-container">
      <va-progress-circle indeterminate />
    </div>

    <div v-else-if="template" class="template-detail">
      <!-- Header -->
      <div class="page-header">
        <div class="header-left">
          <va-button preset="secondary" size="small" @click="goBack">
            <va-icon name="arrow_back" class="mr-2" />
            Back
          </va-button>
          <h1 class="page-title">{{ template.name }}</h1>
          <va-badge v-if="template.is_public" text="Public" color="success" />
          <va-badge v-else text="Private" color="secondary" />
        </div>
        <div class="header-actions">
          <va-button preset="primary" @click="handleUseTemplate">
            <va-icon name="add" class="mr-2" />
            Create Workspace
          </va-button>
          <va-button v-if="authStore.isAdmin" preset="secondary" @click="showEditDialog = true">
            <va-icon name="edit" class="mr-2" />
            Edit
          </va-button>
          <va-button v-if="authStore.isAdmin" preset="secondary" @click="handleClone">
            <va-icon name="content_copy" class="mr-2" />
            Clone
          </va-button>
          <va-button v-if="authStore.isAdmin" preset="danger" @click="showDeleteDialog = true">
            <va-icon name="delete" class="mr-2" />
            Delete
          </va-button>
        </div>
      </div>

      <!-- Info Cards -->
      <div class="info-grid">
        <!-- Basic Info -->
        <va-card>
          <va-card-title>Basic Information</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="info-label">ID:</span>
              <span class="info-value">{{ template.id }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Description:</span>
              <span class="info-value">{{ template.description || 'N/A' }}</span>
            </div>
            <div v-if="template.category" class="info-row">
              <span class="info-label">Category:</span>
              <span class="info-value">{{ template.category }}</span>
            </div>
            <div v-if="template.owner_username" class="info-row">
              <span class="info-label">Owner:</span>
              <span class="info-value">{{ template.owner_username }}</span>
            </div>
            <div v-if="template.organization_name" class="info-row">
              <span class="info-label">Organization:</span>
              <span class="info-value">{{ template.organization_name }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Created:</span>
              <span class="info-value">{{ formatDate(template.created_at) }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Updated:</span>
              <span class="info-value">{{ formatDate(template.updated_at) }}</span>
            </div>
          </va-card-content>
        </va-card>

        <!-- Stats -->
        <va-card>
          <va-card-title>Statistics</va-card-title>
          <va-card-content>
            <div class="info-row">
              <span class="info-label">Usage Count:</span>
              <span class="info-value">{{ template.usage_count || 0 }}</span>
            </div>
            <div class="info-row">
              <span class="info-label">Visibility:</span>
              <span class="info-value">{{ template.is_public ? 'Public' : 'Private' }}</span>
            </div>
          </va-card-content>
        </va-card>
      </div>

      <!-- Docker Image -->
      <va-card class="mt-4">
        <va-card-title>Docker Image</va-card-title>
        <va-card-content>
          <div class="image-display">
            <code>{{ template.image }}</code>
          </div>
        </va-card-content>
      </va-card>

      <!-- Default Resource Limits -->
      <va-card class="mt-4">
        <va-card-title>Default Resource Limits</va-card-title>
        <va-card-content>
          <div class="resources-grid">
            <div class="resource-item">
              <div class="resource-label">CPU</div>
              <div class="resource-value">{{ template.default_cpu_limit || 'Unlimited' }}</div>
            </div>
            <div class="resource-item">
              <div class="resource-label">Memory</div>
              <div class="resource-value">{{ formatMemory(template.default_memory_limit) }}</div>
            </div>
            <div class="resource-item">
              <div class="resource-label">Storage</div>
              <div class="resource-value">{{ formatStorage(template.default_storage_limit) }}</div>
            </div>
          </div>
        </va-card-content>
      </va-card>

      <!-- Ports -->
      <va-card v-if="template.ports && template.ports.length > 0" class="mt-4">
        <va-card-title>Exposed Ports</va-card-title>
        <va-card-content>
          <div class="ports-list">
            <va-chip v-for="port in template.ports" :key="port" color="primary">
              {{ port }}
            </va-chip>
          </div>
        </va-card-content>
      </va-card>

      <!-- Default Environment Variables -->
      <va-card
        v-if="template.default_env_vars && Object.keys(template.default_env_vars).length > 0"
        class="mt-4"
      >
        <va-card-title>Default Environment Variables</va-card-title>
        <va-card-content>
          <va-data-table :items="envVarsList" :columns="envVarsColumns" striped />
        </va-card-content>
      </va-card>

      <!-- Tags -->
      <va-card v-if="template.tags && template.tags.length > 0" class="mt-4">
        <va-card-title>Tags</va-card-title>
        <va-card-content>
          <div class="tags-list">
            <va-chip v-for="tag in template.tags" :key="tag" color="primary" outline>
              {{ tag }}
            </va-chip>
          </div>
        </va-card-content>
      </va-card>
    </div>

    <div v-else class="error-state">
      <p>Template not found</p>
      <va-button @click="goBack">Go Back</va-button>
    </div>

    <!-- Edit Dialog Placeholder -->
    <va-modal v-model="showEditDialog" title="Edit Template" size="large">
      <p>Template edit form will be implemented here</p>
      <template #footer>
        <va-button preset="secondary" @click="showEditDialog = false">Cancel</va-button>
        <va-button @click="handleUpdate">Save</va-button>
      </template>
    </va-modal>

    <!-- Delete Confirmation -->
    <va-modal
      v-model="showDeleteDialog"
      title="Delete Template"
      message="Are you sure you want to delete this template? This action cannot be undone."
      ok-text="Delete"
      cancel-text="Cancel"
      @ok="confirmDelete"
    />

    <!-- Clone Dialog -->
    <va-modal v-model="showCloneDialog" title="Clone Template" size="medium">
      <va-input
        v-model="cloneName"
        label="New Template Name"
        placeholder="Enter name for cloned template"
      />
      <template #footer>
        <va-button preset="secondary" @click="showCloneDialog = false">Cancel</va-button>
        <va-button @click="confirmClone" :loading="templateStore.loading">Clone</va-button>
      </template>
    </va-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTemplateStore } from '@/stores/template'
import { formatDate as formatDateUtil } from '@/utils/date'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const templateStore = useTemplateStore()

// State
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const showCloneDialog = ref(false)
const cloneName = ref('')

// Computed
const templateId = computed(() => route.params.id as string)
const template = computed(() => templateStore.currentTemplate)

const envVarsList = computed(() => {
  if (!template.value?.default_env_vars) return []
  return Object.entries(template.value.default_env_vars).map(([key, value]) => ({
    key,
    value,
  }))
})

const envVarsColumns = [
  { key: 'key', label: 'Key' },
  { key: 'value', label: 'Value' },
]

// Methods
function formatDate(date: string): string {
  return formatDateUtil(date)
}

function formatMemory(bytes: number | undefined): string {
  if (!bytes) return 'Unlimited'
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(2)} GB`
}

function formatStorage(bytes: number | undefined): string {
  if (!bytes) return 'Unlimited'
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(2)} GB`
}

function goBack(): void {
  router.push({ name: 'templates' })
}

function handleUseTemplate(): void {
  router.push({
    name: 'workspaces',
    query: { template: templateId.value },
  })
}

function handleUpdate(): void {
  // Placeholder
  console.log('Update template')
  showEditDialog.value = false
}

async function confirmDelete(): Promise<void> {
  try {
    await templateStore.deleteTemplate(templateId.value)
    router.push({ name: 'templates' })
  } catch (error) {
    console.error('Failed to delete template:', error)
  }
}

function handleClone(): void {
  if (template.value) {
    cloneName.value = `${template.value.name} (Copy)`
    showCloneDialog.value = true
  }
}

async function confirmClone(): Promise<void> {
  if (cloneName.value) {
    try {
      await templateStore.cloneTemplate(templateId.value, cloneName.value)
      showCloneDialog.value = false
      cloneName.value = ''
      router.push({ name: 'templates' })
    } catch (error) {
      console.error('Failed to clone template:', error)
    }
  }
}

async function loadTemplate(): Promise<void> {
  try {
    await templateStore.fetchTemplate(templateId.value)
  } catch (error) {
    console.error('Failed to load template:', error)
  }
}

// Lifecycle
onMounted(() => {
  loadTemplate()
})
</script>

<style lang="scss" scoped>
.template-detail-page {
  .loading-container,
  .error-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 400px;
    gap: 1rem;
  }

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    flex-wrap: wrap;
    gap: 1rem;

    .header-left {
      display: flex;
      align-items: center;
      gap: 1rem;
      flex-wrap: wrap;

      .page-title {
        margin: 0;
        font-size: 2rem;
        font-weight: 600;
      }
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }
  }

  .info-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .info-row {
    display: flex;
    justify-content: space-between;
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--va-background-border);

    &:last-child {
      border-bottom: none;
    }

    .info-label {
      font-weight: 500;
      color: var(--va-text-secondary);
    }

    .info-value {
      text-align: right;
      word-break: break-word;
      max-width: 60%;
    }
  }

  .image-display {
    padding: 1rem;
    background-color: var(--va-background-element);
    border-radius: 4px;

    code {
      font-size: 1rem;
      word-break: break-all;
    }
  }

  .resources-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 1.5rem;

    .resource-item {
      text-align: center;
      padding: 1rem;
      background-color: var(--va-background-element);
      border-radius: 8px;

      .resource-label {
        font-size: 0.875rem;
        color: var(--va-text-secondary);
        margin-bottom: 0.5rem;
      }

      .resource-value {
        font-size: 1.25rem;
        font-weight: 600;
      }
    }
  }

  .ports-list,
  .tags-list {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .mt-4 {
    margin-top: 1.5rem;
  }
}
</style>
