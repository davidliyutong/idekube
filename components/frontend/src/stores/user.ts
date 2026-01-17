import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
import { useNotification } from '@/utils/notification'

export interface User {
  id: string
  username: string
  email: string
  full_name?: string
  avatar?: string
  status: 'active' | 'inactive' | 'suspended'
  roles: string[]
  organization_ids?: string[]
  is_admin?: boolean
  created_at: string
  updated_at: string
  last_login?: string
}

export interface UserListParams {
  page?: number
  page_size?: number
  search?: string
  status?: string
  role?: string
  sort_by?: string
  sort_order?: 'asc' | 'desc'
}

export interface UserCreateData {
  username: string
  email: string
  full_name?: string
  password?: string
  roles?: string[]
}

export interface UserUpdateData {
  full_name?: string
  email?: string
  roles?: string[]
  status?: string
}

export const useUserStore = defineStore('user', () => {
  const notification = useNotification()

  // State
  const users = ref<User[]>([])
  const currentUser = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const total = ref(0)
  const currentPage = ref(1)
  const pageSize = ref(20)

  // Available roles for assignment
  const availableRoles = ref<string[]>([
    'admin',
    'operator',
    'developer',
    'viewer',
  ])

  // Getters
  const userById = computed(() => {
    return (id: string) => users.value.find((u) => u.id === id)
  })

  const adminUsers = computed(() => {
    return users.value.filter((u) => u.is_admin)
  })

  const activeUsers = computed(() => {
    return users.value.filter((u) => u.status === 'active')
  })

  // Actions
  async function fetchUsers(params: UserListParams = {}): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<{
        users: User[]
        total: number
        page: number
        page_size: number
      }>('/v1/users', { params })

      users.value = response.data.users
      total.value = response.data.total
      currentPage.value = response.data.page
      pageSize.value = response.data.page_size
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch users'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function fetchUser(id: string): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get<User>(`/v1/users/${id}`)
      currentUser.value = response.data

      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = response.data
      }

      return response.data
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch user'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createUser(data: UserCreateData): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.post<User>('/v1/users', data)
      const newUser = response.data

      users.value.unshift(newUser)
      total.value++
      notification.success('User created successfully')

      return newUser
    } catch (err: any) {
      error.value = err.message || 'Failed to create user'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateUser(id: string, data: UserUpdateData): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<User>(`/v1/users/${id}`, data)
      const updatedUser = response.data

      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = updatedUser
      }

      if (currentUser.value?.id === id) {
        currentUser.value = updatedUser
      }

      notification.success('User updated successfully')
      return updatedUser
    } catch (err: any) {
      error.value = err.message || 'Failed to update user'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deleteUser(id: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete(`/v1/users/${id}`)

      users.value = users.value.filter((u) => u.id !== id)
      total.value--

      if (currentUser.value?.id === id) {
        currentUser.value = null
      }

      notification.success('User deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete user'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function updateUserRoles(id: string, roles: string[]): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<User>(`/v1/users/${id}`, { roles })
      const updatedUser = response.data

      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = updatedUser
      }

      if (currentUser.value?.id === id) {
        currentUser.value = updatedUser
      }

      notification.success('User roles updated successfully')
      return updatedUser
    } catch (err: any) {
      error.value = err.message || 'Failed to update user roles'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function setUserAdmin(id: string, isAdmin: boolean): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<User>(`/v1/users/${id}`, {
        is_admin: isAdmin,
      })
      const updatedUser = response.data

      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = updatedUser
      }

      if (currentUser.value?.id === id) {
        currentUser.value = updatedUser
      }

      const message = isAdmin ? 'User promoted to admin' : 'Admin privileges removed'
      notification.success(message)
      return updatedUser
    } catch (err: any) {
      error.value = err.message || 'Failed to update user admin status'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  async function changeUserStatus(id: string, status: string): Promise<User> {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.patch<User>(`/v1/users/${id}`, { status })
      const updatedUser = response.data

      const index = users.value.findIndex((u) => u.id === id)
      if (index !== -1) {
        users.value[index] = updatedUser
      }

      if (currentUser.value?.id === id) {
        currentUser.value = updatedUser
      }

      notification.success('User status updated successfully')
      return updatedUser
    } catch (err: any) {
      error.value = err.message || 'Failed to update user status'
      notification.error(error.value)
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearError(): void {
    error.value = null
  }

  function setCurrentUser(user: User | null): void {
    currentUser.value = user
  }

  return {
    // State
    users,
    currentUser,
    loading,
    error,
    total,
    currentPage,
    pageSize,
    availableRoles,

    // Getters
    userById,
    adminUsers,
    activeUsers,

    // Actions
    fetchUsers,
    fetchUser,
    createUser,
    updateUser,
    deleteUser,
    updateUserRoles,
    setUserAdmin,
    changeUserStatus,
    clearError,
    setCurrentUser,
  }
})
