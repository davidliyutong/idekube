<template>
  <div class="login-page">
    <va-form @submit.prevent="handleLogin">
      <va-input
        v-model="credentials.username"
        label="Username or Email"
        placeholder="Enter your username or email"
        :rules="[(v) => !!v || 'Username is required']"
        class="mb-4"
      />
      
      <va-input
        v-model="credentials.password"
        type="password"
        label="Password"
        placeholder="Enter your password"
        :rules="[(v) => !!v || 'Password is required']"
        class="mb-4"
      />

      <div class="login-page__actions">
        <va-button type="submit" :loading="authStore.loading" class="mb-2">
          Login
        </va-button>

        <div class="login-page__links">
          <router-link :to="{ name: 'forgot-password' }">
            Forgot password?
          </router-link>
          <router-link :to="{ name: 'register' }">
            Create account
          </router-link>
        </div>
      </div>
    </va-form>

    <va-divider class="my-4">OR</va-divider>

    <div v-if="oidcProviders.length > 0" class="login-page__oidc">
      <va-button
        v-for="provider in oidcProviders"
        :key="provider.id"
        preset="secondary"
        class="mb-2"
        @click="handleOIDCLogin(provider.id)"
      >
        <va-icon name="login" class="mr-2" />
        Login with {{ provider.name }}
      </va-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@stores/auth'
import { useNotification } from '@utils/notification'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const notification = useNotification()

const credentials = ref({
  username: '',
  password: '',
})

const oidcProviders = ref<any[]>([])

async function handleLogin() {
  try {
    await authStore.login(credentials.value)
    notification.success('Login successful')
    
    const redirect = route.query.redirect as string
    router.push(redirect || { name: 'dashboard' })
  } catch (error) {
    notification.error('Login failed. Please check your credentials.')
  }
}

function handleOIDCLogin(providerId: string) {
  // Redirect to OIDC provider login endpoint
  window.location.href = `/api/v1/auth/oidc/${providerId}/login`
}

async function loadOIDCProviders() {
  try {
    // This would fetch from /api/v1/auth/oidc/providers
    // For now, just placeholder
    oidcProviders.value = []
  } catch (error) {
    console.error('Failed to load OIDC providers:', error)
  }
}

onMounted(() => {
  loadOIDCProviders()
})
</script>

<style lang="scss" scoped>
.login-page {
  &__actions {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    .va-button {
      width: 100%;
    }
  }

  &__links {
    display: flex;
    justify-content: space-between;
    font-size: 0.875rem;

    a {
      color: var(--va-primary);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  &__oidc {
    display: flex;
    flex-direction: column;

    .va-button {
      width: 100%;
    }
  }
}
</style>
