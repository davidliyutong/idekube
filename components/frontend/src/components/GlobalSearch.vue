<template>
  <va-modal
    v-model="isOpen"
    title="Global Search"
    size="large"
    hide-default-actions
  >
    <div class="global-search">
      <div class="search-input">
        <va-input
          id="global-search-input"
          v-model="query"
          placeholder="Search workspaces, templates, users, organizations, volumes..."
          @keyup.enter="triggerSearch"
        >
          <template #prependInner>
            <va-icon name="search" />
          </template>
          <template #appendInner>
            <va-chip size="small" color="secondary">⌘K</va-chip>
          </template>
        </va-input>
      </div>

      <div class="filters">
        <va-tabs v-model="tabIndex">
          <va-tab v-for="tab in tabs" :key="tab.value">{{ tab.label }}</va-tab>
        </va-tabs>
      </div>

      <div class="results">
        <div v-if="loading" class="loading">
          <va-progress-circle indeterminate />
        </div>

        <div v-else-if="filteredResults.length === 0" class="empty">
          <va-icon name="search" size="large" color="secondary" />
          <p>Type to search across resources</p>
        </div>

        <va-list v-else class="result-list">
          <va-list-item
            v-for="item in filteredResults"
            :key="item.type + '-' + item.id"
            @click="navigate(item)"
          >
            <va-list-item-section>
              <va-icon :name="item.icon || 'search'" class="mr-2" />
              <div class="item-texts">
                <div class="title">
                  {{ item.title }}
                </div>
                <div class="subtitle">
                  <va-badge :text="item.type" color="info" />
                  <span v-if="item.subtitle" class="ml-2">{{ item.subtitle }}</span>
                </div>
              </div>
            </va-list-item-section>
          </va-list-item>
        </va-list>
      </div>

      <div class="helper">
        <va-chip size="small" class="mr-2">Enter: open</va-chip>
        <va-chip size="small" class="mr-2">Esc: close</va-chip>
        <va-chip size="small">? : shortcuts</va-chip>
      </div>
    </div>
  </va-modal>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useSearchStore } from '../stores/search'
import { useShortcutsStore } from '../stores/shortcuts'

const router = useRouter()
const searchStore = useSearchStore()
const shortcutsStore = useShortcutsStore()

const isOpen = computed({
  get: () => searchStore.isOpen,
  set: (v: boolean) => (v ? searchStore.open() : searchStore.close()),
})

const query = computed({
  get: () => searchStore.query,
  set: (v: string) => searchStore.setQuery(v),
})

const loading = computed(() => searchStore.loading)
const filteredResults = computed(() => searchStore.filteredResults)

const tabs = [
  { label: 'All', value: 'all' },
  { label: 'Workspaces', value: 'workspace' },
  { label: 'Templates', value: 'template' },
  { label: 'Users', value: 'user' },
  { label: 'Organizations', value: 'organization' },
  { label: 'Volumes', value: 'volume' },
]
const tabIndex = ref(0)

watch(tabIndex, (i) => {
  const v = tabs[i]?.value || 'all'
  // @ts-ignore
  searchStore.setFilter(v)
})

function triggerSearch() {
  searchStore.search()
}

function navigate(item: any) {
  router.push({ name: item.routeName, params: item.routeParams })
  searchStore.close()
  searchStore.clear()
}

// Optional: live search on query change (debounced)
let timer: number | null = null
watch(() => query.value, (q) => {
  if (timer) window.clearTimeout(timer)
  timer = window.setTimeout(() => searchStore.search(q), 250)
})

// Listen Esc to close
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape' && searchStore.isOpen) {
    searchStore.close()
  } else if (e.key === '?' || (e.shiftKey && e.key === '/')) {
    shortcutsStore.openHelp()
  }
}

window.addEventListener('keydown', onKeydown)
</script>

<style scoped lang="scss">
.global-search {
  display: grid;
  grid-template-rows: auto auto 1fr auto;
  gap: 1rem;

  .search-input {
    position: relative;
  }

  .results {
    min-height: 220px;

    .loading,
    .empty {
      display: flex;
      align-items: center;
      justify-content: center;
      color: var(--va-textColorSecondary);
      gap: 0.5rem;
      padding: 2rem 0;
      flex-direction: column;
    }

    .result-list {
      max-height: 340px;
      overflow: auto;
    }

    .item-texts {
      display: flex;
      flex-direction: column;
      .title { font-weight: 600; }
      .subtitle { font-size: 0.85rem; color: var(--va-textColorSecondary); }
    }
  }

  .helper {
    display: flex;
    align-items: center;
  }
}

.ml-2 { margin-left: 0.5rem; }
.mr-2 { margin-right: 0.5rem; }
</style>
