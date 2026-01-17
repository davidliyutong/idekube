import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiClient } from '@/api'
import { useNotification, type UseNotificationReturn } from '@/composables/useNotification'

export interface APIKey {
  id: number
  name: string
  scopes: string[]
  created_at: string
  expires_at: string | null
  last_used_at: string | null
  user_id: number
}

export interface CreateAPIKeyRequest {
  name: string
  scopes: string[]
  expires_at?: number
}

export const useAPIKeyStore = defineStore('apiKey', () => {
  const notification = useNotification()
  const { showNotification } = notification

  // State
  const apiKeys = ref<APIKey[]>([])
  const currentKey = ref<string | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed
  const activeKeys = computed(() => {
    return apiKeys.value.filter((key) => {
      if (!key.expires_at) return true
      return new Date(key.expires_at) > new Date()
    })
  })

  const expiredKeys = computed(() => {
    return apiKeys.value.filter((key) => {
      if (!key.expires_at) return false
      return new Date(key.expires_at) <= new Date()
    })
  })

  // Actions
  async function fetchAPIKeys() {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<{ data: APIKey[] }>('/api-keys')
      apiKeys.value = response.data.data || []
      return apiKeys.value
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch API keys'
      error.value = message
      showNotification({
        title: 'Error',
        message,
        color: 'danger',
      })
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createAPIKey(data: CreateAPIKeyRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.post<{ data: { key: string; api_key: APIKey } }>(
        '/api-keys',
        data
      )
      
      // Store the generated key temporarily
      currentKey.value = response.data.data.key
      
      await fetchAPIKeys()
      
      showNotification({
        title: 'Success',
        message: 'API key created successfully',
        color: 'success',
      })

      return response.data.data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create API key'
      error.value = message
      showNotification({
        title: 'Error',
        message,
        color: 'danger',
      })
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteAPIKey(id: number) {
    loading.value = true
    error.value = null

    try {
      await apiClient.delete(`/api-keys/${id}`)
      
      apiKeys.value = apiKeys.value.filter((key) => key.id !== id)
      
      showNotification({
        title: 'Success',
        message: 'API key deleted successfully',
        color: 'success',
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete API key'
      error.value = message
      showNotification({
        title: 'Error',
        message,
        color: 'danger',
      })
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearCurrentKey() {
    currentKey.value = null
  }

  return {
    // State
    apiKeys,
    currentKey,
    loading,
    error,

    // Computed
    activeKeys,
    expiredKeys,

    // Actions
    fetchAPIKeys,
    createAPIKey,
    deleteAPIKey,
    clearCurrentKey,
  }
})
