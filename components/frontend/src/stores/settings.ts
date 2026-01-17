import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { apiClient } from '@/api'
import { useNotification } from '@/composables/useNotification'

// Type definitions
export interface SettingResponse {
  key: string
  value: string
  category: string
  description: string
  value_type: string
  is_public: boolean
  updated_at: string
}

export interface GetSettingsResponse {
  settings: SettingResponse[]
}

export interface BatchUpdateSettingsRequest {
  settings: Record<string, string>
}

export interface UpdateSettingRequest {
  value: string
}

// Settings categories
export const SETTING_CATEGORIES = {
  system: 'System Configuration',
  email: 'Email Settings',
  oidc: 'OIDC Configuration',
  storage: 'Storage Settings',
  security: 'Security Settings',
  monitoring: 'Monitoring Settings',
  quota: 'Quota Settings',
  other: 'Other Settings',
} as const

export type SettingCategory = keyof typeof SETTING_CATEGORIES

export const useSettingsStore = defineStore('settings', () => {
  const { showNotification } = useNotification()

  // State
  const settings = ref<SettingResponse[]>([])
  const publicSettings = ref<SettingResponse[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Computed properties
  const settingsByCategory = computed(() => {
    const grouped: Record<string, SettingResponse[]> = {}

    settings.value.forEach((setting) => {
      const category = setting.category || 'other'
      if (!grouped[category]) {
        grouped[category] = []
      }
      grouped[category].push(setting)
    })

    return grouped
  })

  const getSetting = (key: string) => {
    return settings.value.find((s) => s.key === key)
  }

  const getSettingValue = (key: string): string | null => {
    return getSetting(key)?.value || null
  }

  const getPublicSetting = (key: string) => {
    return publicSettings.value.find((s) => s.key === key)
  }

  const getPublicSettingValue = (key: string): string | null => {
    return getPublicSetting(key)?.value || null
  }

  // Actions

  /**
   * Fetch all system settings (admin only)
   */
  async function fetchSettings() {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<GetSettingsResponse>('/api/v1/settings')
      settings.value = response.data.settings || []
      return settings.value
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch settings'
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

  /**
   * Fetch public settings (no authentication required)
   */
  async function fetchPublicSettings() {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<GetSettingsResponse>('/api/v1/settings/public')
      publicSettings.value = response.data.settings || []
      return publicSettings.value
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch public settings'
      error.value = message
      // Don't show error notification for public settings since they're optional
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * Get a single setting by key (admin only)
   */
  async function fetchSetting(key: string) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.get<SettingResponse>(`/api/v1/settings/${key}`)
      
      // Update or add the setting in the array
      const index = settings.value.findIndex((s) => s.key === key)
      if (index >= 0) {
        settings.value[index] = response.data
      } else {
        settings.value.push(response.data)
      }

      return response.data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to fetch setting'
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

  /**
   * Update a single setting by key (admin only)
   */
  async function updateSetting(key: string, value: string) {
    loading.value = true
    error.value = null

    try {
      const response = await apiClient.put<SettingResponse>(`/api/v1/settings/${key}`, {
        value,
      } as UpdateSettingRequest)

      // Update the setting in the array
      const index = settings.value.findIndex((s) => s.key === key)
      if (index >= 0) {
        settings.value[index] = response.data
      } else {
        settings.value.push(response.data)
      }

      showNotification({
        title: 'Success',
        message: `Setting "${key}" updated successfully`,
        color: 'success',
      })

      return response.data
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update setting'
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

  /**
   * Update multiple settings at once (admin only)
   */
  async function batchUpdateSettings(settingsData: Record<string, string>) {
    loading.value = true
    error.value = null

    try {
      await apiClient.put('/api/v1/settings', {
        settings: settingsData,
      } as BatchUpdateSettingsRequest)

      // Refresh all settings
      await fetchSettings()

      showNotification({
        title: 'Success',
        message: 'Settings updated successfully',
        color: 'success',
      })
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update settings'
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

  /**
   * Clear all settings from memory
   */
  function clearSettings() {
    settings.value = []
    publicSettings.value = []
    error.value = null
  }

  return {
    // State
    settings,
    publicSettings,
    loading,
    error,

    // Computed
    settingsByCategory,
    getSetting,
    getSettingValue,
    getPublicSetting,
    getPublicSettingValue,

    // Actions
    fetchSettings,
    fetchPublicSettings,
    fetchSetting,
    updateSetting,
    batchUpdateSettings,
    clearSettings,
  }
})
