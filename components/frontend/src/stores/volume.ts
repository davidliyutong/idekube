import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
import { useNotification } from '@/utils/notification'

export interface Volume {
  id: string
  name: string
  description?: string
  size: number
  owner_id: string
  owner_username?: string
  organization_id?: string
  organization_name?: string
  attached_to?: string[]
  status: 'available' | 'in-use' | 'error'
  labels?: Record<string, string>
  created_at: string
  updated_at: string
}

export interface VolumeListParams {
  page?: number
  page_size?: number
  search?: string
  owner_id?: string
  organization_id?: string
  status?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface VolumeCreateData {
  name: string
  description?: string
  size: number
  organization_id?: string
  labels?: Record<string, string>
}

export interface VolumeUpdateData {
  name?: string
  description?: string
  size?: number
  labels?: Record<string, string>
}

export const useVolumeStore = defineStore('volume', () => {
  const notification = useNotification()

  // State
  const volumes = ref<Volume[]>([])
  const currentVolume = ref<Volume | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // Getters
  const volumeById = computed(() => {
    return (id: string) => volumes.value.find((v) => v.id === id)
  })

  const availableVolumes = computed(() => {
    return volumes.value.filter((v) => v.status === 'available')
  })

  // Actions
  async function fetchVolumes(params: VolumeListParams = {}): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<{
        volumes: Volume[]
        total: number
        page: number
        page_size: number
      }>('/v1/volumes', { params })

      volumes.value = response.data.volumes
      total.value = response.data.total
      currentPage.value = response.data.page
      pageSize.value = response.data.page_size
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch volumes'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchVolume(id: string): Promise<Volume> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<Volume>(`/v1/volumes/${id}`)
      currentVolume.value = response.data

      const index = volumes.value.findIndex((v) => v.id === id)
      if (index !== -1) {
        volumes.value[index] = response.data
      }

      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createVolume(data: VolumeCreateData): Promise<Volume> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Volume>('/v1/volumes', data)
      const newVolume = response.data

      volumes.value.unshift(newVolume)
      total.value++
      notification.success('Volume created successfully')

      return newVolume
    } catch (err: any) {
      error.value = err.message || 'Failed to create volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateVolume(id: string, data: VolumeUpdateData): Promise<Volume> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<Volume>(`/v1/volumes/${id}`, data)
      const updatedVolume = response.data

      const index = volumes.value.findIndex((v) => v.id === id)
      if (index !== -1) {
        volumes.value[index] = updatedVolume
      }

      if (currentVolume.value?.id === id) {
        currentVolume.value = updatedVolume
      }

      notification.success('Volume updated successfully')
      return updatedVolume
    } catch (err: any) {
      error.value = err.message || 'Failed to update volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteVolume(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/volumes/${id}`)

      volumes.value = volumes.value.filter((v) => v.id !== id)
      total.value--

      if (currentVolume.value?.id === id) {
        currentVolume.value = null
      }

      notification.success('Volume deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function syncVolume(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.post(`/v1/volumes/${id}/sync`)
      notification.success('Volume sync initiated')
      await fetchVolume(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to sync volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  function setCurrentVolume(volume: Volume | null): void {
    currentVolume.value = volume
  }

  return {
    // State
    volumes,
    currentVolume,
    loading,
    error,
    total,
    currentPage,
    pageSize,

    // Getters
    volumeById,
    availableVolumes,

    // Actions
    fetchVolumes,
    fetchVolume,
    createVolume,
    updateVolume,
    deleteVolume,
    syncVolume,
    clearError,
    setCurrentVolume,
  }
})
