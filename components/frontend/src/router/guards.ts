import type { Router } from 'vue-router'
import { useAuthStore } from '@stores/auth'

export function setupRouterGuards(router: Router) {
  router.beforeEach((to, _from, next) => {
    const authStore = useAuthStore()

    // Check if route requires authentication
    if (to.meta.requiresAuth && !authStore.isAuthenticated) {
      next({
        name: 'login',
        query: { redirect: to.fullPath },
      })
      return
    }

    // Check if route requires guest (not authenticated)
    if (to.meta.requiresGuest && authStore.isAuthenticated) {
      next({ name: 'dashboard' })
      return
    }

    // Check if route requires admin
    if (to.meta.requiresAdmin && !authStore.isAdmin) {
      next({ name: 'dashboard' })
      return
    }

    next()
  })
}
