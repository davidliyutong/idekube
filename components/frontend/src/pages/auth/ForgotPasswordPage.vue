<template>
  <div class="forgot-password-page">
    <p class="mb-4">
      Enter your email address and we'll send you a link to reset your password.
    </p>

    <va-form @submit.prevent="handleSubmit">
      <va-input
        v-model="email"
        type="email"
        label="Email"
        placeholder="Enter your email"
        :rules="[(v) => !!v || 'Email is required', (v) => /.+@.+\..+/.test(v) || 'Email must be valid']"
        class="mb-4"
      />

      <va-button type="submit" :loading="loading" class="mb-4">
        Send Reset Link
      </va-button>

      <div class="forgot-password-page__links">
        <router-link :to="{ name: 'login' }">
          Back to login
        </router-link>
      </div>
    </va-form>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@stores/auth'
import { useNotification } from '@utils/notification'

const router = useRouter()
const authStore = useAuthStore()
const notification = useNotification()

const loading = ref(false)
const email = ref('')

async function handleSubmit() {
  loading.value = true
  try {
    await authStore.requestPasswordReset(email.value)
    notification.success('Password reset link sent! Please check your email.')
    router.push({ name: 'login' })
  } catch (error) {
    notification.error('Failed to send reset link. Please try again.')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.forgot-password-page {
  &__links {
    text-align: center;
    font-size: 0.875rem;

    a {
      color: var(--va-primary);
      text-decoration: none;

      &:hover {
        text-decoration: underline;
      }
    }
  }

  .va-button {
    width: 100%;
  }
}
</style>
