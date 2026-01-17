import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
import { useNotification } from '@/utils/notification'

export interface Template {
  id: string
  name: string
  description?: string
  category?: string
  image: string
  default_cpu_limit?: number
  default_memory_limit?: number
  default_storage_limit?: number
  default_env_vars?: Record<string, string>
  ports?: number[]
  is_public: boolean
  owner_id?: string
  owner_username?: string
  organization_id?: string
  organization_name?: string
  tags?: string[]
  usage_count?: number
  created_at: string
  updated_at: string
}

export interface TemplateListParams {
  page?: number
  page_size?: number
  search?: string
  category?: string
  is_public?: boolean
  owner_id?: string
  organization_id?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface TemplateCreateData {
  name: string
  description?: string
  category?: string
  image: string
  default_cpu_limit?: number
  default_memory_limit?: number
  default_storage_limit?: number
  default_env_vars?: Record<string, string>
  ports?: number[]
  is_public: boolean
  organization_id?: string
  tags?: string[]
}

export interface TemplateUpdateData {
  name?: string
  description?: string
  category?: string
  image?: string
  default_cpu_limit?: number
  default_memory_limit?: number
  default_storage_limit?: number
  default_env_vars?: Record<string, string>
  ports?: number[]
  is_public?: boolean
  tags?: string[]
}

export const useTemplateStore = defineStore('template', () => {
  const notification = useNotification()

  // State
  const templates = ref<Template[]>([])
  const currentTemplate = ref<Template | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // Cache for quick access
  const templateCache = ref<Map<string, Template>>(new Map())

  // Getters
  const templateById = computed(() => {
    return (id: string) => {
      return templateCache.value.get(id) || templates.value.find((t) => t.id === id)
    }
  })

  const publicTemplates = computed(() => {
    return templates.value.filter((t) => t.is_public)
  })

  const templatesByCategory = computed(() => {
    return (category: string) => templates.value.filter((t) => t.category === category)
  })

  const categories = computed(() => {
    const cats = new Set(templates.value.map((t) => t.category).filter(Boolean))
    return Array.from(cats) as string[]
  })

  // Actions
  async function fetchTemplates(params: TemplateListParams = {}): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<{
        templates: Template[]
        total: number
        page: number
        page_size: number
      }>('/v1/templates', { params })

      templates.value = response.data.templates
      total.value = response.data.total
      currentPage.value = response.data.page
      pageSize.value = response.data.page_size

      // Update cache
      response.data.templates.forEach((template) => {
        templateCache.value.set(template.id, template)
      })
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch templates'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchTemplate(id: string): Promise<Template> {
    // Check cache first
    const cached = templateCache.value.get(id)
    if (cached) {
      currentTemplate.value = cached
      return cached
    }

    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<Template>(`/v1/templates/${id}`)
      currentTemplate.value = response.data

      // Update cache
      templateCache.value.set(id, response.data)

      // Update in list if exists
      const index = templates.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        templates.value[index] = response.data
      }

      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch template'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createTemplate(data: TemplateCreateData): Promise<Template> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Template>('/v1/templates', data)
      const newTemplate = response.data

      templates.value.unshift(newTemplate)
      total.value++
      templateCache.value.set(newTemplate.id, newTemplate)
      notification.success('Template created successfully')

      return newTemplate
    } catch (err: any) {
      error.value = err.message || 'Failed to create template'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateTemplate(id: string, data: TemplateUpdateData): Promise<Template> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<Template>(`/v1/templates/${id}`, data)
      const updatedTemplate = response.data

      // Update in list
      const index = templates.value.findIndex((t) => t.id === id)
      if (index !== -1) {
        templates.value[index] = updatedTemplate
      }

      // Update cache
      templateCache.value.set(id, updatedTemplate)

      if (currentTemplate.value?.id === id) {
        currentTemplate.value = updatedTemplate
      }

      notification.success('Template updated successfully')
      return updatedTemplate
    } catch (err: any) {
      error.value = err.message || 'Failed to update template'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteTemplate(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/templates/${id}`)

      // Remove from list
      templates.value = templates.value.filter((t) => t.id !== id)
      total.value--

      // Remove from cache
      templateCache.value.delete(id)

      if (currentTemplate.value?.id === id) {
        currentTemplate.value = null
      }

      notification.success('Template deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete template'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function cloneTemplate(id: string, newName: string): Promise<Template> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Template>(`/v1/templates/${id}/clone`, {
        name: newName,
      })
      const clonedTemplate = response.data

      templates.value.unshift(clonedTemplate)
      total.value++
      templateCache.value.set(clonedTemplate.id, clonedTemplate)
      notification.success('Template cloned successfully')

      return clonedTemplate
    } catch (err: any) {
      error.value = err.message || 'Failed to clone template'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  function setCurrentTemplate(template: Template | null): void {
    currentTemplate.value = template
  }

  function clearCache(): void {
    templateCache.value.clear()
  }

  return {
    // State
    templates,
    currentTemplate,
    loading,
    error,
    total,
    currentPage,
    pageSize,

    // Getters
    templateById,
    publicTemplates,
    templatesByCategory,
    categories,

    // Actions
    fetchTemplates,
    fetchTemplate,
    createTemplate,
    updateTemplate,
    deleteTemplate,
    cloneTemplate,
    clearError,
    setCurrentTemplate,
    clearCache,
  }
})
