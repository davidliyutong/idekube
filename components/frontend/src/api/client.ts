import axios, { AxiosError } from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig } from 'axios'
import { useAuthStore } from '@stores/auth'
import { useNotification } from '@utils/notification'

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const authStore = useAuthStore()
    const token = authStore.token

    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// Response interceptor
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  async (error: AxiosError) => {
    const authStore = useAuthStore()
    const notification = useNotification()
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }

    // Handle 401 Unauthorized
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        // Try to refresh token
        const refreshed = await authStore.refreshTokenAction()
        
        if (refreshed && originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${authStore.token}`
          return apiClient(originalRequest)
        }
      } catch (refreshError) {
        // Refresh failed, logout user
        authStore.logout()
        notification.error('Session expired. Please login again.')
        return Promise.reject(refreshError)
      }
    }

    // Handle other errors
    if (error.response) {
      const message = (error.response.data as any)?.message || error.message
      
      switch (error.response.status) {
        case 403:
          notification.error('Access denied')
          break
        case 404:
          notification.error('Resource not found')
          break
        case 500:
          notification.error('Server error: ' + message)
          break
        default:
          notification.error(message)
      }
    } else if (error.request) {
      notification.error('Network error. Please check your connection.')
    } else {
      notification.error('Request failed: ' + error.message)
    }

    return Promise.reject(error)
  }
)

export default apiClient
