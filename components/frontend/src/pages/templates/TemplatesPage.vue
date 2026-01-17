<template>
  <div class="templates-page">
    <div class="page-header">
      <h1 class="page-title">Templates</h1>
      <va-button v-if="authStore.isAdmin" @click="showCreateDialog = true">
        <va-icon name="add" class="mr-2" />
        New Template
      </va-button>
    </div>

    <!-- Filters -->
    <va-card class="mb-4">
      <va-card-content>
        <div class="filters">
          <va-input
            v-model="searchQuery"
            placeholder="Search templates..."
            class="filter-input"
            @update:model-value="handleSearch"
          >
            <template #prepend>
              <va-icon name="search" />
            </template>
          </va-input>

          <va-select
            v-model="categoryFilter"
            label="Category"
            :options="categoryOptions"
            class="filter-select"
            @update:model-value="handleFilterChange"
          />

          <va-select
            v-model="visibilityFilter"
            label="Visibility"
            :options="visibilityOptions"
            class="filter-select"
            @update:model-value="handleFilterChange"
          />

          <va-button preset="secondary" @click="resetFilters">
            <va-icon name="refresh" class="mr-2" />
            Reset
          </va-button>
        </div>
      </va-card-content>
    </va-card>

    <!-- Templates Grid -->
    <va-card :loading="templateStore.loading">
      <va-card-content>
        <div v-if="templateStore.templates.length === 0" class="empty-state">
          <va-icon name="inbox" size="4rem" color="secondary" />
          <p>No templates found</p>
          <va-button v-if="authStore.isAdmin" @click="showCreateDialog = true">
            Create First Template
          </va-button>
        </div>

        <div v-else class="templates-grid">
          <va-card
            v-for="template in templateStore.templates"
            :key="template.id"
            class="template-card"
            @click="goToDetail(template.id)"
          >
            <va-card-content>
              <div class="template-header">
                <div class="template-icon">
                  <va-icon name="description" size="2rem" />
                </div>
                <div class="template-badges">
                  <va-badge v-if="template.is_public" text="Public" color="success" />
                  <va-badge v-else text="Private" color="secondary" />
                </div>
              </div>

              <h3 class="template-name">{{ template.name }}</h3>
              <p class="template-description">
                {{ template.description || 'No description' }}
              </p>

              <div class="template-meta">
                <div v-if="template.category" class="meta-item">
                  <va-icon name="category" size="small" />
                  <span>{{ template.category }}</span>
                </div>
                <div v-if="template.usage_count !== undefined" class="meta-item">
                  <va-icon name="people" size="small" />
                  <span>{{ template.usage_count }} uses</span>
                </div>
              </div>

              <div class="template-image">
                <code>{{ template.image }}</code>
              </div>

              <div v-if="template.tags && template.tags.length > 0" class="template-tags">
                <va-chip
                  v-for="tag in template.tags.slice(0, 3)"
                  :key="tag"
                  size="small"
                  color="primary"
                  outline
                >
                  {{ tag }}
                </va-chip>
                <va-chip v-if="template.tags.length > 3" size="small" color="secondary">
                  +{{ template.tags.length - 3 }}
                </va-chip>
              </div>

              <div class="template-actions">
                <va-button
                  size="small"
                  preset="primary"
                  @click.stop="handleUseTemplate(template)"
                >
                  Use Template
                </va-button>
                <va-button
                  v-if="authStore.isAdmin"
                  size="small"
                  preset="secondary"
                  @click.stop="handleEdit(template)"
                >
                  <va-icon name="edit" />
                </va-button>
                <va-button
                  v-if="authStore.isAdmin"
                  size="small"
                  preset="secondary"
                  @click.stop="handleClone(template)"
                >
                  <va-icon name="content_copy" />
                </va-button>
              </div>
            </va-card-content>
          </va-card>
        </div>

        <!-- Pagination -->
        <div v-if="totalPages > 1" class="pagination">
          <va-pagination
            v-model="currentPage"
            :pages="totalPages"
            :visible-pages="5"
            @update:model-value="handlePageChange"
          />
          <div class="pagination-info">
            Showing {{ startIndex }} - {{ endIndex }} of {{ templateStore.total }} templates
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Create/Edit Dialog Placeholder -->
    <va-modal v-model="showCreateDialog" title="Create Template" size="large">
      <p>Template creation form will be implemented here</p>
      <template #footer>
        <va-button preset="secondary" @click="showCreateDialog = false">Cancel</va-button>
        <va-button @click="handleCreate">Create</va-button>
      </template>
    </va-modal>

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
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTemplateStore } from '@/stores/template'
import type { Template } from '@/stores/template'

const router = useRouter()
const authStore = useAuthStore()
const templateStore = useTemplateStore()

// State
const showCreateDialog = ref(false)
const showCloneDialog = ref(false)
const templateToClone = ref<Template | null>(null)
const cloneName = ref('')
const searchQuery = ref('')
const categoryFilter = ref('')
const visibilityFilter = ref('')
const currentPage = ref(1)
const pageSize = 20

// Options
const categoryOptions = computed(() => [
  { text: 'All', value: '' },
  ...templateStore.categories.map((cat) => ({ text: cat, value: cat })),
])

const visibilityOptions = [
  { text: 'All', value: '' },
  { text: 'Public', value: 'true' },
  { text: 'Private', value: 'false' },
]

// Computed
const totalPages = computed(() => Math.ceil(templateStore.total / pageSize))

const startIndex = computed(() => (currentPage.value - 1) * pageSize + 1)

const endIndex = computed(() => {
  const end = currentPage.value * pageSize
  return end > templateStore.total ? templateStore.total : end
})

// Methods
async function loadTemplates(): Promise<void> {
  await templateStore.fetchTemplates({
    page: currentPage.value,
    page_size: pageSize,
    search: searchQuery.value || undefined,
    category: categoryFilter.value || undefined,
    is_public: visibilityFilter.value ? visibilityFilter.value === 'true' : undefined,
  })
}

function handleSearch(): void {
  currentPage.value = 1
  loadTemplates()
}

function handleFilterChange(): void {
  currentPage.value = 1
  loadTemplates()
}

function resetFilters(): void {
  searchQuery.value = ''
  categoryFilter.value = ''
  visibilityFilter.value = ''
  currentPage.value = 1
  loadTemplates()
}

function handlePageChange(): void {
  loadTemplates()
}

function goToDetail(id: string): void {
  router.push({ name: 'template-detail', params: { id } })
}

function handleUseTemplate(template: Template): void {
  router.push({
    name: 'workspaces',
    query: { template: template.id },
  })
}

function handleEdit(template: Template): void {
  router.push({ name: 'template-detail', params: { id: template.id } })
}

function handleClone(template: Template): void {
  templateToClone.value = template
  cloneName.value = `${template.name} (Copy)`
  showCloneDialog.value = true
}

async function confirmClone(): Promise<void> {
  if (templateToClone.value && cloneName.value) {
    try {
      await templateStore.cloneTemplate(templateToClone.value.id, cloneName.value)
      showCloneDialog.value = false
      templateToClone.value = null
      cloneName.value = ''
      loadTemplates()
    } catch (error) {
      console.error('Failed to clone template:', error)
    }
  }
}

function handleCreate(): void {
  // Placeholder
  console.log('Create template')
  showCreateDialog.value = false
}

// Lifecycle
onMounted(() => {
  loadTemplates()
})
</script>

<style lang="scss" scoped>
.templates-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .filters {
    display: flex;
    gap: 1rem;
    align-items: flex-end;

    .filter-input {
      flex: 1;
      min-width: 200px;
    }

    .filter-select {
      min-width: 150px;
    }
  }

  .mb-4 {
    margin-bottom: 1.5rem;
  }

  .empty-state {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 3rem;
    text-align: center;
    color: var(--va-text-secondary);

    p {
      margin: 1rem 0;
      font-size: 1.1rem;
    }
  }

  .templates-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
    gap: 1.5rem;
  }

  .template-card {
    cursor: pointer;
    transition: transform 0.2s, box-shadow 0.2s;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 16px rgba(0, 0, 0, 0.1);
    }

    .template-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;

      .template-icon {
        color: var(--va-primary);
      }

      .template-badges {
        display: flex;
        gap: 0.5rem;
      }
    }

    .template-name {
      margin: 0 0 0.5rem 0;
      font-size: 1.25rem;
      font-weight: 600;
    }

    .template-description {
      margin: 0 0 1rem 0;
      color: var(--va-text-secondary);
      font-size: 0.875rem;
      line-height: 1.4;
      display: -webkit-box;
      -webkit-line-clamp: 2;
      -webkit-box-orient: vertical;
      overflow: hidden;
    }

    .template-meta {
      display: flex;
      gap: 1rem;
      margin-bottom: 1rem;

      .meta-item {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        font-size: 0.875rem;
        color: var(--va-text-secondary);
      }
    }

    .template-image {
      margin-bottom: 1rem;
      padding: 0.5rem;
      background-color: var(--va-background-element);
      border-radius: 4px;

      code {
        font-size: 0.75rem;
        word-break: break-all;
      }
    }

    .template-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
      margin-bottom: 1rem;
    }

    .template-actions {
      display: flex;
      gap: 0.5rem;
      padding-top: 1rem;
      border-top: 1px solid var(--va-background-border);
    }
  }

  .pagination {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 1.5rem;

    .pagination-info {
      color: var(--va-text-secondary);
      font-size: 0.875rem;
    }
  }
}
</style>
