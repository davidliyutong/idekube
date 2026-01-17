import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import apiClient from '@/api/client'
// Notification optional; avoid hard dependency to keep store lightweight

export type ResourceType = 'workspace' | 'template' | 'user' | 'organization' | 'volume'

export interface SearchResultItem {
  type: ResourceType
  id: string
  title: string
  subtitle?: string
  icon?: string
  routeName: string
  routeParams?: Record<string, string>
}

export const useSearchStore = defineStore('search', () => {
  // const { showError } = useNotification()

  const isOpen = ref(false)
  const query = ref('')
  const loading = ref(false)
  const filter: { value: 'all' | ResourceType } = { value: 'all' }
  const results = ref<SearchResultItem[]>([])

  const filteredResults = computed(() => {
    if (filter.value === 'all') return results.value
    return results.value.filter(r => r.type === filter.value)
  })

  function open() {
    isOpen.value = true
    // small delay to allow modal to render
    setTimeout(() => {
      const el = document.querySelector<HTMLInputElement>('#global-search-input')
      el?.focus()
    }, 50)
  }

  function close() {
    isOpen.value = false
  }

  function setQuery(value: string) {
    query.value = value
  }

  function setFilter(value: 'all' | ResourceType) {
    filter.value = value
  }

  async function search(q?: string) {
    const term = (q ?? query.value).trim()
    if (!term) {
      results.value = []
      return
    }

    loading.value = true
    try {
      const limit = 5
      const params = (extra: Record<string, any> = {}) => ({ page_size: limit, search: term, ...extra })

      const requests = [
        apiClient.get('/v1/workspaces', { params: params() }).catch(() => ({ data: { items: [] } })),
        apiClient.get('/v1/templates', { params: params() }).catch(() => ({ data: { items: [] } })),
        apiClient.get('/v1/users', { params: params() }).catch(() => ({ data: { items: [] } })),
        apiClient.get('/v1/organizations', { params: params() }).catch(() => ({ data: { items: [] } })),
        apiClient.get('/v1/volumes', { params: params() }).catch(() => ({ data: { items: [] } })),
      ]

      const [wsRes, tplRes, usrRes, orgRes, volRes] = await Promise.all(requests)

      const mapItems = (items: any[], type: ResourceType, routeName: string, titleKey: string, subtitleKey?: string, icon?: string): SearchResultItem[] => {
        return (items || []).map((it: any) => ({
          type,
          id: String(it.id),
          title: String(it[titleKey] ?? it.name ?? it.title ?? it.id),
          subtitle: subtitleKey ? String(it[subtitleKey] ?? '') : undefined,
          icon,
          routeName,
          routeParams: { id: String(it.id) },
        }))
      }

      const all: SearchResultItem[] = [
        ...mapItems(wsRes.data.items || wsRes.data || [], 'workspace', 'workspace-detail', 'name', 'template_name', 'computer'),
        ...mapItems(tplRes.data.items || tplRes.data || [], 'template', 'template-detail', 'name', 'version', 'layers'),
        ...mapItems(usrRes.data.items || usrRes.data || [], 'user', 'user-detail', 'username', 'full_name', 'person'),
        ...mapItems(orgRes.data.items || orgRes.data || [], 'organization', 'organization-detail', 'name', undefined, 'business'),
        ...mapItems(volRes.data.items || volRes.data || [], 'volume', 'volume-detail', 'name', 'size', 'storage'),
      ]

      results.value = all
    } catch (err: any) {
      // showError('Global search failed')
      console.error('Global search error:', err)
    } finally {
      loading.value = false
    }
  }

  function clear() {
    query.value = ''
    results.value = []
  }

  return {
    isOpen,
    query,
    loading,
    results,
    filteredResults,
    open,
    close,
    setQuery,
    setFilter,
    search,
    clear,
  }
})
