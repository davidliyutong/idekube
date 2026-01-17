<template>
  <router-view />
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuthStore } from '@stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  // Initialize auth state from storage
  authStore.initializeAuth()
})
</script>

<style lang="scss">
#app {
  height: 100vh;
  width: 100vw;
  overflow: hidden;
}
</style>
