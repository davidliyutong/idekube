import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { storage } from '@utils/storage'
import { isTokenExpired } from '@utils/auth'
import apiClient from '@/api/client'

export interface User {
  id: string
  username: string
  email: string
  role: string
  full_name?: string
  is_active?: boolean
  is_verified?: boolean
  mfa_enabled?: boolean
  created_at?: string
  updated_at?: string
}

export interface LoginCredentials {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  refresh_token: string
  user: User
}

export const useAuthStore = defineStore('auth', () => {
  // State
  const token = ref<string | null>(storage.getToken())
  const refreshToken = ref<string | null>(storage.getRefreshToken())
  const user = ref<User | null>(storage.getUser<User>())
  const loading = ref(false)

  // Getters
  const isAuthenticated = computed(() => {
    return !!token.value && !!user.value && !isTokenExpired(token.value)
  })

  const isAdmin = computed(() => {
    return user.value?.role === 'admin'
  })

  const userId = computed(() => user.value?.id)
  const username = computed(() => user.value?.username)

  // Actions
  function initializeAuth() {
    const storedToken = storage.getToken()
    const storedRefreshToken = storage.getRefreshToken()
    const storedUser = storage.getUser<User>()

    if (storedToken && storedUser) {
      if (isTokenExpired(storedToken)) {
        // Token expired, try to refresh
        refreshTokenAction()
      } else {
        token.value = storedToken
        refreshToken.value = storedRefreshToken
        user.value = storedUser
      }
    }
  }

  async function login(credentials: LoginCredentials): Promise<void> {
    loading.value = true
    try {
      const response = await apiClient.post<LoginResponse>('/v1/auth/login', credentials)
      const data = response.data

      setAuth(data.token, data.refresh_token, data.user)
    } catch (error) {
      clearAuth()
      throw error
    } finally {
      loading.value = false
    }
  }

  async function logout(): Promise<void> {
    try {
      await apiClient.post('/v1/auth/logout')
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearAuth()
    }
  }

  async function refreshTokenAction(): Promise<boolean> {
    const currentRefreshToken = refreshToken.value
    if (!currentRefreshToken) {
      clearAuth()
      return false
    }

    try {
      const response = await apiClient.post<LoginResponse>('/v1/auth/refresh', {
        refresh_token: currentRefreshToken,
      })
      const data = response.data

      setAuth(data.token, data.refresh_token, data.user)
      return true
    } catch (error) {
      console.error('Token refresh error:', error)
      clearAuth()
      return false
    }
  }

  async function register(data: {
    username: string
    email: string
    password: string
    full_name?: string
  }): Promise<void> {
    loading.value = true
    try {
      await apiClient.post('/v1/auth/register', data)
    } finally {
      loading.value = false
    }
  }

  async function requestPasswordReset(email: string): Promise<void> {
    loading.value = true
    try {
      await apiClient.post('/v1/auth/request-password-reset', { email })
    } finally {
      loading.value = false
    }
  }

  async function resetPassword(token: string, password: string): Promise<void> {
    loading.value = true
    try {
      await apiClient.post('/v1/auth/reset-password', { token, password })
    } finally {
      loading.value = false
    }
  }

  async function verifyEmail(token: string): Promise<void> {
    loading.value = true
    try {
      await apiClient.get('/v1/auth/verify-email', { params: { token } })
      if (user.value) {
        user.value.is_verified = true
        storage.setUser(user.value)
      }
    } finally {
      loading.value = false
    }
  }

  async function updateProfile(data: Partial<User>): Promise<void> {
    loading.value = true
    try {
      const response = await apiClient.put<User>('/v1/users/me', data)
      user.value = response.data
      storage.setUser(response.data)
    } finally {
      loading.value = false
    }
  }

  async function changePassword(oldPassword: string, newPassword: string): Promise<void> {
    loading.value = true
    try {
      await apiClient.post('/v1/users/me/password', {
        old_password: oldPassword,
        new_password: newPassword,
      })
    } finally {
      loading.value = false
    }
  }

  function setAuth(newToken: string, newRefreshToken: string, newUser: User) {
    token.value = newToken
    refreshToken.value = newRefreshToken
    user.value = newUser

    storage.setToken(newToken)
    storage.setRefreshToken(newRefreshToken)
    storage.setUser(newUser)
  }

  function clearAuth() {
    token.value = null
    refreshToken.value = null
    user.value = null

    storage.clear()
  }

  return {
    // State
    token,
    refreshToken,
    user,
    loading,
    // Getters
    isAuthenticated,
    isAdmin,
    userId,
    username,
    // Actions
    initializeAuth,
    login,
    logout,
    refreshTokenAction,
    register,
    requestPasswordReset,
    resetPassword,
    verifyEmail,
    updateProfile,
    changePassword,
  }
})
