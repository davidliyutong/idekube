import { defineStore } from 'pinia'
import { ref } from 'vue'
import { apiClient } from '@/api'
import { useNotification } from '@/composables/useNotification'

export interface Webhook {
  id: number
  url: string
  events: string[]
  enabled: boolean
  created_at: string
  updated_at: string
  last_triggered_at: string | null
  user_id: number
}

export interface CreateWebhookRequest {
  url: string
  events: string[]
  enabled?: boolean
}

export interface UpdateWebhookRequest {
  url?: string
  events?: string[]
  enabled?: boolean
}

export const useWebhookStore = defineStore('webhook', () => {
  const { showNotification } = useNotification()

  // State
  const webhooks = ref<Webhook[]>([])
  const currentWebhook = ref<Webhook | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Actions
  async function fetchWebhooks() {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<{ data: Webhook[] }>('/webhooks')
      webhooks.value = response.data.data || []
      return webhooks.value
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch webhooks'
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

  async function fetchWebhook(id: number) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<{ data: Webhook }>(`/webhooks/${id}`)
      currentWebhook.value = response.data.data
      return currentWebhook.value
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch webhook'
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

  async function createWebhook(data: CreateWebhookRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.post<{ data: Webhook }>('/webhooks', data)
      
      await fetchWebhooks()
      
      showNotification({
        title: 'Success',
        message: 'Webhook created successfully',
        color: 'success',
      })

      return response.data.data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create webhook'
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

  async function updateWebhook(id: number, data: UpdateWebhookRequest) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.put<{ data: Webhook }>(`/webhooks/${id}`, data)
      
      const index = webhooks.value.findIndex((w) => w.id === id)
      if (index >= 0) {
        webhooks.value[index] = response.data.data
      }
      
      showNotification({
        title: 'Success',
        message: 'Webhook updated successfully',
        color: 'success',
      })

      return response.data.data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update webhook'
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

  async function deleteWebhook(id: number) {
    loading.value = true
    error.value = null

    try {
      await apiClient.delete(`/webhooks/${id}`)
      
      webhooks.value = webhooks.value.filter((w) => w.id !== id)
      
      showNotification({
        title: 'Success',
        message: 'Webhook deleted successfully',
        color: 'success',
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete webhook'
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

  async function testWebhook(id: number) {
    loading.value = true
    error.value = null

    try {
      await apiClient.post(`/webhooks/${id}/test`)
      
      showNotification({
        title: 'Success',
        message: 'Test event sent to webhook',
        color: 'success',
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to test webhook'
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

  return {
    // State
    webhooks,
    currentWebhook,
    loading,
    error,

    // Actions
    fetchWebhooks,
    fetchWebhook,
    createWebhook,
    updateWebhook,
    deleteWebhook,
    testWebhook,
  }
})
