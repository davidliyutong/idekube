import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
import { useNotification } from '@/utils/notification'

export interface Organization {
  id: string
  name: string
  description?: string
  owner_id: string
  owner_username?: string
  members: OrganizationMember[]
  admins: string[]
  workspace_quota?: number
  storage_quota?: number
  status: 'active' | 'inactive' | 'suspended'
  created_at: string
  updated_at: string
}

export interface OrganizationMember {
  id: string
  username: string
  email: string
  role: 'admin' | 'member'
  joined_at: string
}

export interface OrganizationListParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface OrganizationCreateData {
  name: string
  description?: string
  workspace_quota?: number
  storage_quota?: number
}

export interface OrganizationUpdateData {
  name?: string
  description?: string
  workspace_quota?: number
  storage_quota?: number
  status?: string
}

export const useOrganizationStore = defineStore('organization', () => {
  const notification = useNotification()

  // State
  const organizations = ref<Organization[]>([])
  const currentOrganization = ref<Organization | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // Getters
  const organizationById = computed(() => {
    return (id: string) => organizations.value.find((o) => o.id === id)
  })

  const activeOrganizations = computed(() => {
    return organizations.value.filter((o) => o.status === 'active')
  })

  // Actions
  async function fetchOrganizations(
    params: OrganizationListParams = {}
  ): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<{
        organizations: Organization[]
        total: number
        page: number
        page_size: number
      }>('/v1/organizations', { params })

      organizations.value = response.data.organizations
      total.value = response.data.total
      currentPage.value = response.data.page
      pageSize.value = response.data.page_size
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch organizations'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchOrganization(id: string): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<Organization>(`/v1/organizations/${id}`)
      currentOrganization.value = response.data

      const index = organizations.value.findIndex((o) => o.id === id)
      if (index !== -1) {
        organizations.value[index] = response.data
      }

      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch organization'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createOrganization(data: OrganizationCreateData): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Organization>('/v1/organizations', data)
      const newOrganization = response.data

      organizations.value.unshift(newOrganization)
      total.value++
      notification.success('Organization created successfully')

      return newOrganization
    } catch (err: any) {
      error.value = err.message || 'Failed to create organization'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateOrganization(
    id: string,
    data: OrganizationUpdateData
  ): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<Organization>(
        `/v1/organizations/${id}`,
        data
      )
      const updatedOrganization = response.data

      const index = organizations.value.findIndex((o) => o.id === id)
      if (index !== -1) {
        organizations.value[index] = updatedOrganization
      }

      if (currentOrganization.value?.id === id) {
        currentOrganization.value = updatedOrganization
      }

      notification.success('Organization updated successfully')
      return updatedOrganization
    } catch (err: any) {
      error.value = err.message || 'Failed to update organization'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteOrganization(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/organizations/${id}`)

      organizations.value = organizations.value.filter((o) => o.id !== id)
      total.value--

      if (currentOrganization.value?.id === id) {
        currentOrganization.value = null
      }

      notification.success('Organization deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete organization'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function addMember(organizationId: string, userId: string): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<Organization>(
        `/v1/organizations/${organizationId}/members`,
        { user_id: userId }
      )
      const updatedOrganization = response.data

      const index = organizations.value.findIndex((o) => o.id === organizationId)
      if (index !== -1) {
        organizations.value[index] = updatedOrganization
      }

      if (currentOrganization.value?.id === organizationId) {
        currentOrganization.value = updatedOrganization
      }

      notification.success('Member added successfully')
      return updatedOrganization
    } catch (err: any) {
      error.value = err.message || 'Failed to add member'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function removeMember(organizationId: string, userId: string): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.delete<Organization>(
        `/v1/organizations/${organizationId}/members/${userId}`
      )
      const updatedOrganization = response.data

      const index = organizations.value.findIndex((o) => o.id === organizationId)
      if (index !== -1) {
        organizations.value[index] = updatedOrganization
      }

      if (currentOrganization.value?.id === organizationId) {
        currentOrganization.value = updatedOrganization
      }

      notification.success('Member removed successfully')
      return updatedOrganization
    } catch (err: any) {
      error.value = err.message || 'Failed to remove member'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function setMemberAdmin(
    organizationId: string,
    userId: string,
    isAdmin: boolean
  ): Promise<Organization> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<Organization>(
        `/v1/organizations/${organizationId}/members/${userId}`,
        { is_admin: isAdmin }
      )
      const updatedOrganization = response.data

      const index = organizations.value.findIndex((o) => o.id === organizationId)
      if (index !== -1) {
        organizations.value[index] = updatedOrganization
      }

      if (currentOrganization.value?.id === organizationId) {
        currentOrganization.value = updatedOrganization
      }

      const message = isAdmin ? 'Member promoted to admin' : 'Admin privileges removed'
      notification.success(message)
      return updatedOrganization
    } catch (err: any) {
      error.value = err.message || 'Failed to update member admin status'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  function setCurrentOrganization(organization: Organization | null): void {
    currentOrganization.value = organization
  }

  return {
    // State
    organizations,
    currentOrganization,
    loading,
    error,
    total,
    currentPage,
    pageSize,

    // Getters
    organizationById,
    activeOrganizations,

    // Actions
    fetchOrganizations,
    fetchOrganization,
    createOrganization,
    updateOrganization,
    deleteOrganization,
    addMember,
    removeMember,
    setMemberAdmin,
    clearError,
    setCurrentOrganization,
  }
})
