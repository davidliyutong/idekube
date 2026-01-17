import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
import { useNotification } from '@/utils/notification'

export interface Workspace {
  id: string
  name: string
  description?: string
  template_id: string
  template_name?: string
  owner_id: string
  owner_username?: string
  organization_id?: string
  organization_name?: string
  status: 'pending' | 'running' | 'stopped' | 'error' | 'deleting'
  cpu_limit?: number
  memory_limit?: number
  storage_limit?: number
  image?: string
  port?: number
  env_vars?: Record<string, string>
  volumes?: Array<{
    id: string
    name: string
    mount_path: string
  }>
  created_at: string
  updated_at: string
  started_at?: string
  stopped_at?: string
}

export interface WorkspaceListParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
  template_id?: string
  owner_id?: string
  organization_id?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface WorkspaceCreateData {
  name: string
  description?: string
  template_id: string
  organization_id?: string
  cpu_limit?: number
  memory_limit?: number
  storage_limit?: number
  env_vars?: Record<string, string>
  volumes?: string[]
}

export interface WorkspaceUpdateData {
  name?: string
  description?: string
  cpu_limit?: number
  memory_limit?: number
  storage_limit?: number
  env_vars?: Record<string, string>
}

export const useWorkspaceStore = defineStore('workspace', () => {
  const notification = useNotification()

  // State
  const workspaces = ref<Workspace[]>([])
  const currentWorkspace = ref<Workspace | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // Getters
  const workspaceById = computed(() => {
    return (id: string) => workspaces.value.find((w) => w.id === id)
  })

  const runningWorkspaces = computed(() => {
    return workspaces.value.filter((w) => w.status === 'running')
  })

  const stoppedWorkspaces = computed(() => {
    return workspaces.value.filter((w) => w.status === 'stopped')
  })

  // Actions
  async function fetchWorkspaces(params: WorkspaceListParams = {}): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<{
        workspaces: Workspace[]
        total: number
        page: number
        page_size: number
      }>('/v1/workspaces', { params })

      workspaces.value = response.data.workspaces
      total.value = response.data.total
      currentPage.value = response.data.page
      pageSize.value = response.data.page_size
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch workspaces'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchWorkspace(id: string): Promise<Workspace> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<Workspace>(`/v1/workspaces/${id}`)
      currentWorkspace.value = response.data

      // Update in list if exists
      const index = workspaces.value.findIndex((w) => w.id === id)
      if (index !== -1) {
        workspaces.value[index] = response.data
      }

      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createWorkspace(data: WorkspaceCreateData): Promise<Workspace> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Workspace>('/v1/workspaces', data)
      const newWorkspace = response.data

      workspaces.value.unshift(newWorkspace)
      total.value++
      notification.success('Workspace created successfully')

      return newWorkspace
    } catch (err: any) {
      error.value = err.message || 'Failed to create workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateWorkspace(id: string, data: WorkspaceUpdateData): Promise<Workspace> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<Workspace>(`/v1/workspaces/${id}`, data)
      const updatedWorkspace = response.data

      // Update in list
      const index = workspaces.value.findIndex((w) => w.id === id)
      if (index !== -1) {
        workspaces.value[index] = updatedWorkspace
      }

      if (currentWorkspace.value?.id === id) {
        currentWorkspace.value = updatedWorkspace
      }

      notification.success('Workspace updated successfully')
      return updatedWorkspace
    } catch (err: any) {
      error.value = err.message || 'Failed to update workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteWorkspace(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/workspaces/${id}`)

      // Remove from list
      workspaces.value = workspaces.value.filter((w) => w.id !== id)
      total.value--

      if (currentWorkspace.value?.id === id) {
        currentWorkspace.value = null
      }

      notification.success('Workspace deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function startWorkspace(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.post(`/v1/workspaces/${id}/start`)

      // Update status optimistically
      const workspace = workspaceById.value(id)
      if (workspace) {
        workspace.status = 'running'
      }

      notification.success('Workspace started successfully')

      // Refresh workspace details
      await fetchWorkspace(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to start workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function stopWorkspace(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.post(`/v1/workspaces/${id}/stop`)

      // Update status optimistically
      const workspace = workspaceById.value(id)
      if (workspace) {
        workspace.status = 'stopped'
      }

      notification.success('Workspace stopped successfully')

      // Refresh workspace details
      await fetchWorkspace(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to stop workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function transferWorkspace(id: string, targetUserId: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.post(`/v1/workspaces/${id}/transfer`, {
        target_user_id: targetUserId,
      })

      notification.success('Workspace transfer requested')

      // Refresh workspace details
      await fetchWorkspace(id)
    } catch (err: any) {
      error.value = err.message || 'Failed to transfer workspace'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function attachVolume(workspaceId: string, volumeId: string, mountPath: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.post(`/v1/workspaces/${workspaceId}/volumes`, {
        volume_id: volumeId,
        mount_path: mountPath,
      })

      notification.success('Volume attached successfully')

      // Refresh workspace details
      await fetchWorkspace(workspaceId)
    } catch (err: any) {
      error.value = err.message || 'Failed to attach volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function detachVolume(workspaceId: string, volumeId: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/workspaces/${workspaceId}/volumes/${volumeId}`)

      notification.success('Volume detached successfully')

      // Refresh workspace details
      await fetchWorkspace(workspaceId)
    } catch (err: any) {
      error.value = err.message || 'Failed to detach volume'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  function setCurrentWorkspace(workspace: Workspace | null): void {
    currentWorkspace.value = workspace
  }

  // Auto-refresh for workspace status (optional)
  let refreshInterval: ReturnType<typeof setInterval> | null = null

  function startAutoRefresh(intervalMs: number = 10000): void {
    stopAutoRefresh()
    refreshInterval = setInterval(() => {
      if (workspaces.value.length > 0) {
        fetchWorkspaces({
          page: currentPage.value,
          page_size: pageSize.value,
        })
      }
    }, intervalMs)
  }

  function stopAutoRefresh(): void {
    if (refreshInterval) {
      clearInterval(refreshInterval)
      refreshInterval = null
    }
  }

  return {
    // State
    workspaces,
    currentWorkspace,
    loading,
    error,
    total,
    currentPage,
    pageSize,

    // Getters
    workspaceById,
    runningWorkspaces,
    stoppedWorkspaces,

    // Actions
    fetchWorkspaces,
    fetchWorkspace,
    createWorkspace,
    updateWorkspace,
    deleteWorkspace,
    startWorkspace,
    stopWorkspace,
    transferWorkspace,
    attachVolume,
    detachVolume,
    clearError,
    setCurrentWorkspace,
    startAutoRefresh,
    stopAutoRefresh,
  }
})
