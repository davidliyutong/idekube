<template>
  <div class="auth-layout">
    <div class="auth-layout__content">
      <div class="auth-layout__card">
        <div class="auth-layout__header">
          <h1 class="auth-layout__title">IDEKube</h1>
          <p class="auth-layout__subtitle">Cloud IDE Management Platform</p>
        </div>
        <router-view />
      </div>
    </div>
    <div class="auth-layout__footer">
      <p>&copy; {{ currentYear }} IDEKube. All rights reserved.</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const currentYear = computed(() => new Date().getFullYear())
</script>

<style lang="scss" scoped>
.auth-layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: linear-gradient(135deg, var(--va-primary) 0%, var(--va-info) 100%);

  &__content {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
  }

  &__card {
    background: var(--va-background-element);
    border-radius: 12px;
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    padding: 3rem;
    max-width: 480px;
    width: 100%;
  }

  &__header {
    text-align: center;
    margin-bottom: 2rem;
  }

  &__title {
    font-size: 2rem;
    font-weight: 700;
    color: var(--va-primary);
    margin: 0 0 0.5rem 0;
  }

  &__subtitle {
    font-size: 1rem;
    color: var(--va-text-secondary);
    margin: 0;
  }

  &__footer {
    padding: 1.5rem;
    text-align: center;
    color: rgba(255, 255, 255, 0.8);
    font-size: 0.875rem;

    p {
      margin: 0;
    }
  }
}

@media (max-width: 640px) {
  .auth-layout__card {
    padding: 2rem 1.5rem;
  }

  .auth-layout__title {
    font-size: 1.5rem;
  }
}
</style>
