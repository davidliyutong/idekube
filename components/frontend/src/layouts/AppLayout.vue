<template>
  <va-layout :top="{ fixed: true, order: 2 }" :left="{ fixed: true, absolute: false, order: 1 }">
    <template #top>
      <va-navbar class="app-navbar">
        <template #left>
          <va-button preset="secondary" icon="menu" @click="toggleSidebar" />
          <span class="app-navbar__title">IDEKube</span>
        </template>
        <template #right>
          <va-button preset="secondary" @click="openSearch" class="mr-2">
            <va-icon name="search" class="mr-1" />
            Search
          </va-button>
          <va-button preset="secondary" @click="openShortcuts" class="mr-2">
            <va-icon name="keyboard" class="mr-1" />
            Shortcuts
          </va-button>
          <va-dropdown placement="bottom-end">
            <template #anchor>
              <va-button preset="secondary">
                <va-avatar size="small" :color="userColor">
                  {{ userInitials }}
                </va-avatar>
                <span class="ml-2">{{ user?.username }}</span>
              </va-button>
            </template>
            <va-dropdown-content>
              <va-list>
                <va-list-item @click="goToProfile">
                  <va-icon name="person" class="mr-2" />
                  Profile
                </va-list-item>
                <va-list-item @click="goToAPIKeys">
                  <va-icon name="key" class="mr-2" />
                  API Keys
                </va-list-item>
                <va-list-separator />
                <va-list-item @click="handleLogout">
                  <va-icon name="logout" class="mr-2" />
                  Logout
                </va-list-item>
              </va-list>
            </va-dropdown-content>
          </va-dropdown>
        </template>
      </va-navbar>
    </template>

    <template #left>
      <va-sidebar v-model="isSidebarVisible" :width="sidebarWidth" :minimized="isSidebarMinimized">
        <va-sidebar-item
          v-for="item in menuItems"
          :key="item.name"
          :to="{ name: item.name }"
          :active="isActiveRoute(item.name)"
        >
          <va-sidebar-item-content>
            <va-icon :name="item.icon" />
            <va-sidebar-item-title>
              {{ item.title }}
            </va-sidebar-item-title>
          </va-sidebar-item-content>
        </va-sidebar-item>
      </va-sidebar>
    </template>

    <template #content>
      <main class="app-layout__content">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
        <GlobalSearch />
        <ShortcutsHelpModal />
      </main>
    </template>
  </va-layout>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { defineAsyncComponent } from 'vue'
const GlobalSearch = defineAsyncComponent(() => import('../components/GlobalSearch.vue'))
const ShortcutsHelpModal = defineAsyncComponent(() => import('../components/ShortcutsHelpModal.vue'))
import { useSearchStore } from '../stores/search'
import { useShortcutsStore } from '../stores/shortcuts'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const searchStore = useSearchStore()
const shortcutsStore = useShortcutsStore()

const isSidebarVisible = ref(true)
const isSidebarMinimized = ref(false)
const sidebarWidth = '260px'

const user = computed(() => authStore.user)
const userInitials = computed(() => {
  if (!user.value) return '?'
  const parts = user.value.full_name?.split(' ') || [user.value.username]
  return parts.map(p => p[0]).join('').toUpperCase().slice(0, 2)
})
const userColor = computed(() => {
  const colors = ['primary', 'success', 'danger', 'warning', 'info']
  const index = (user.value?.id.charCodeAt(0) || 0) % colors.length
  return colors[index]
})

interface MenuItem {
  name: string
  title: string
  icon: string
  adminOnly?: boolean
}

const allMenuItems: MenuItem[] = [
  { name: 'dashboard', title: 'Dashboard', icon: 'dashboard' },
  { name: 'workspaces', title: 'Workspaces', icon: 'computer' },
  { name: 'templates', title: 'Templates', icon: 'layers' },
  { name: 'volumes', title: 'Volumes', icon: 'storage' },
  { name: 'organizations', title: 'Organizations', icon: 'business' },
  { name: 'users', title: 'Users', icon: 'people', adminOnly: true },
  { name: 'settings', title: 'Settings', icon: 'settings', adminOnly: true },
  { name: 'webhooks', title: 'Webhooks', icon: 'webhook', adminOnly: true },
]

const menuItems = computed(() => {
  return allMenuItems.filter(item => !item.adminOnly || authStore.isAdmin)
})

function toggleSidebar() {
  if (window.innerWidth < 768) {
    isSidebarVisible.value = !isSidebarVisible.value
  } else {
    isSidebarMinimized.value = !isSidebarMinimized.value
  }
}

function isActiveRoute(name: string): boolean {
  return route.name === name || route.path.startsWith(`/${name}`)
}

function goToProfile() {
  router.push({ name: 'profile' })
}

function goToAPIKeys() {
  router.push({ name: 'api-keys' })
}

async function handleLogout() {
  await authStore.logout()
  router.push({ name: 'login' })
}

function openSearch() {
  searchStore.open()
}

function openShortcuts() {
  shortcutsStore.openHelp()
}
</script>

<style lang="scss" scoped>
.app-navbar {
  &__title {
    margin-left: 1rem;
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--va-primary);
  }
}

.app-layout {
  &__content {
    padding: 1.5rem;
    min-height: calc(100vh - 64px);
    background: var(--va-background-primary);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

.mr-1 { margin-right: 0.25rem; }
.mr-2 { margin-right: 0.5rem; }
.ml-2 { margin-left: 0.5rem; }
</style>
