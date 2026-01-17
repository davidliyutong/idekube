<template>
  <div class="register-page">
    <va-form @submit.prevent="handleRegister">
      <va-input
        v-model="form.username"
        label="Username"
        placeholder="Choose a username"
        :rules="usernameRules"
        class="mb-4"
      />

      <va-input
        v-model="form.email"
        type="email"
        label="Email"
        placeholder="Enter your email"
        :rules="emailRules"
        class="mb-4"
      />

      <va-input
        v-model="form.full_name"
        label="Full Name"
        placeholder="Enter your full name"
        class="mb-4"
      />

      <va-input
        v-model="form.password"
        type="password"
        label="Password"
        placeholder="Choose a password"
        :rules="passwordRules"
        class="mb-4"
      />

      <va-input
        v-model="form.confirmPassword"
        type="password"
        label="Confirm Password"
        placeholder="Confirm your password"
        :rules="confirmPasswordRules"
        class="mb-4"
      />

      <va-button type="submit" :loading="loading" class="mb-4">
        Register
      </va-button>

      <div class="register-page__links">
        <router-link :to="{ name: 'login' }">
          Already have an account? Login
        </router-link>
      </div>
    </va-form>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@stores/auth'
import { useNotification } from '@utils/notification'

const router = useRouter()
const authStore = useAuthStore()
const notification = useNotification()

const loading = ref(false)
const form = ref({
  username: '',
  email: '',
  full_name: '',
  password: '',
  confirmPassword: '',
})

const usernameRules = [
  (v: string) => !!v || 'Username is required',
  (v: string) => v.length >= 3 || 'Username must be at least 3 characters',
]

const emailRules = [
  (v: string) => !!v || 'Email is required',
  (v: string) => /.+@.+\..+/.test(v) || 'Email must be valid',
]

const passwordRules = [
  (v: string) => !!v || 'Password is required',
  (v: string) => v.length >= 8 || 'Password must be at least 8 characters',
]

const confirmPasswordRules = computed(() => [
  (v: string) => !!v || 'Please confirm your password',
  (v: string) => v === form.value.password || 'Passwords do not match',
])

async function handleRegister() {
  loading.value = true
  try {
    await authStore.register({
      username: form.value.username,
      email: form.value.email,
      full_name: form.value.full_name,
      password: form.value.password,
    })
    
    notification.success('Registration successful! Please check your email to verify your account.')
    router.push({ name: 'login' })
  } catch (error: any) {
    notification.error(error.response?.data?.message || 'Registration failed')
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.register-page {
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
