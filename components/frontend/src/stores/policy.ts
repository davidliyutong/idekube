import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import apiClient from '@/api/client'
import { useNotification } from '@/composables/useNotification'

export interface Policy {
  subject: string
  object: string
  action: string
}

export interface CreatePolicyRequest {
  subject: string
  object: string
  action: string
}

export const usePolicyStore = defineStore('policy', () => {
  const policies = ref<Policy[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  const { showSuccess, showError } = useNotification()

  // Computed
  const policyCount = computed(() => policies.value.length)

  const groupedPolicies = computed(() => {
    const groups: Record<string, Policy[]> = {}
    policies.value.forEach(policy => {
      const key = policy.subject
      if (!groups[key]) {
        groups[key] = []
      }
      groups[key].push(policy)
    })
    return groups
  })

  // Actions
  async function fetchPolicies() {
    loading.value = true
    error.value = null
    try {
      const response = await apiClient.get('/policies')
      // API returns nested array [[subject, object, action], ...]
      policies.value = response.data.map((item: [string, string, string]) => ({
        subject: item[0],
        object: item[1],
        action: item[2],
      }))
    } catch (err: any) {
      error.value = err.message || 'Failed to fetch policies'
      showError('Failed to fetch policies')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function createPolicy(data: CreatePolicyRequest) {
    loading.value = true
    error.value = null
    try {
      await apiClient.post('/policies', data)
      await fetchPolicies()
      showSuccess('Policy created successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to create policy'
      showError('Failed to create policy')
      throw err
    } finally {
      loading.value = false
    }
  }

  async function deletePolicy(subject: string, object: string, action: string) {
    loading.value = true
    error.value = null
    try {
      await apiClient.delete('/policies', {
        data: { subject, object, action }
      })
      await fetchPolicies()
      showSuccess('Policy deleted successfully')
    } catch (err: any) {
      error.value = err.message || 'Failed to delete policy'
      showError('Failed to delete policy')
      throw err
    } finally {
      loading.value = false
    }
  }

  function clearPolicies() {
    policies.value = []
    error.value = null
  }

  return {
    // State
    policies,
    loading,
    error,
    // Computed
    policyCount,
    groupedPolicies,
    // Actions
    fetchPolicies,
    createPolicy,
    deletePolicy,
    clearPolicies,
  }
})
