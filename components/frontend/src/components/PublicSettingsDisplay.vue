<template>
  <div class="public-settings-display">
    <!-- System Information Section -->
    <va-card class="info-card">
      <va-card-title>System Information</va-card-title>
      <va-card-content>
        <div class="info-grid">
          <div class="info-item">
            <label>System Name</label>
            <p>{{ getSettingValue('system.name') || 'IDEKube' }}</p>
          </div>
          <div class="info-item">
            <label>Version</label>
            <p>{{ getSettingValue('system.version') || 'Unknown' }}</p>
          </div>
          <div class="info-item">
            <label>API Base URL</label>
            <p class="mono">{{ getSettingValue('system.api_url') || '-' }}</p>
          </div>
          <div class="info-item">
            <label>Support Email</label>
            <p>
              <a :href="`mailto:${getSettingValue('system.support_email')}`">
                {{ getSettingValue('system.support_email') || 'support@example.com' }}
              </a>
            </p>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- OIDC Providers Section -->
    <va-card v-if="oidcProviders.length > 0" class="providers-card">
      <va-card-title>
        <va-icon name="key" class="mr-2" />
        OIDC Authentication Providers
      </va-card-title>
      <va-card-content>
        <div class="providers-list">
          <div v-for="provider in oidcProviders" :key="provider.id" class="provider-item">
            <div class="provider-header">
              <h4>{{ provider.name }}</h4>
              <va-chip
                v-if="provider.enabled"
                color="success"
                size="small"
              >
                Enabled
              </va-chip>
              <va-chip v-else color="secondary" size="small">
                Disabled
              </va-chip>
            </div>
            <div class="provider-details">
              <div class="detail-row">
                <label>Issuer URL:</label>
                <span class="mono">{{ provider.issuer_url }}</span>
              </div>
              <div class="detail-row">
                <label>Scopes:</label>
                <div class="scopes">
                  <va-chip v-for="scope in provider.scopes" :key="scope" size="small">
                    {{ scope }}
                  </va-chip>
                </div>
              </div>
            </div>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Email Settings Section -->
    <va-card v-if="emailSettingsVisible" class="email-card">
      <va-card-title>
        <va-icon name="email" class="mr-2" />
        Email Configuration
      </va-card-title>
      <va-card-content>
        <div class="email-grid">
          <div class="email-item">
            <label>SMTP Host</label>
            <p class="mono">{{ getSettingValue('email.smtp_host') || '-' }}</p>
          </div>
          <div class="email-item">
            <label>SMTP Port</label>
            <p>{{ getSettingValue('email.smtp_port') || '-' }}</p>
          </div>
          <div class="email-item">
            <label>From Address</label>
            <p>{{ getSettingValue('email.from_address') || '-' }}</p>
          </div>
          <div class="email-item">
            <label>From Name</label>
            <p>{{ getSettingValue('email.from_name') || '-' }}</p>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Features Section -->
    <va-card v-if="featureSettings.length > 0" class="features-card">
      <va-card-title>
        <va-icon name="settings" class="mr-2" />
        Enabled Features
      </va-card-title>
      <va-card-content>
        <div class="features-list">
          <div v-for="feature in featureSettings" :key="feature.key" class="feature-item">
            <va-checkbox
              v-model="featureToggle"
              :disabled="true"
              :option="feature.key"
              class="feature-checkbox"
            >
              {{ formatFeatureName(feature.key) }}
            </va-checkbox>
            <span class="feature-status" :class="getFeatureStatus(feature.value)">
              {{ feature.value === 'true' ? 'Enabled' : 'Disabled' }}
            </span>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Quotas Section -->
    <va-card v-if="quotaSettings.length > 0" class="quotas-card">
      <va-card-title>
        <va-icon name="storage" class="mr-2" />
        Default Quotas
      </va-card-title>
      <va-card-content>
        <div class="quotas-grid">
          <div v-for="quota in quotaSettings" :key="quota.key" class="quota-item">
            <label>{{ formatQuotaName(quota.key) }}</label>
            <p>{{ quota.value }}</p>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Loading State -->
    <div v-if="loading" class="loading-overlay">
      <va-progress-circle indeterminate />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useSettingsStore } from '@/stores/settings'

const settingsStore = useSettingsStore()

// Refs
const featureToggle = ref<string[]>([])

// Computed
const { loading, publicSettings } = settingsStore

const oidcProviders = computed<any[]>(() => {
  // OIDC providers would be fetched from a separate API endpoint
  // This is just a placeholder
  return []
})

// Helper function to get setting value
function getSettingValue(key: string): string | undefined {
  const setting = publicSettings.find((s: any) => s.key === key)
  return setting?.value
}

function getPublicSettingValue(key: string): string | undefined {
  return getSettingValue(key)
}

const emailSettingsVisible = computed(() => {
  return (
    getPublicSettingValue('email.smtp_host') ||
    getPublicSettingValue('email.from_address')
  )
})

const featureSettings = computed(() => {
  return publicSettings.filter((s: any) => s.category === 'features')
})

const quotaSettings = computed(() => {
  return publicSettings.filter((s: any) => s.category === 'quota')
})

// Methods
function getFeatureStatus(value: string) {
  return value === 'true' ? 'enabled' : 'disabled'
}

function formatFeatureName(key: string): string {
  return key
    .replace(/^features\./, '')
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

function formatQuotaName(key: string): string {
  return key
    .replace(/^quota\./, '')
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())
}

// Lifecycle
onMounted(async () => {
  try {
    await settingsStore.fetchPublicSettings()
  } catch (error) {
    console.error('Failed to load public settings:', error)
  }
})
</script>

<style lang="scss" scoped>
.public-settings-display {
  position: relative;
  display: grid;
  grid-template-columns: 1fr;
  gap: 1.5rem;

  .info-card,
  .providers-card,
  .email-card,
  .features-card,
  .quotas-card {
    va-card-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 1.1rem;
      font-weight: 600;
    }
  }

  .info-card {
    .info-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 1.5rem;

      .info-item {
        label {
          display: block;
          font-size: 0.875rem;
          font-weight: 500;
          color: var(--va-textColorSecondary);
          margin-bottom: 0.5rem;
        }

        p {
          margin: 0;
          font-size: 1rem;
          color: var(--va-textColor);

          a {
            color: var(--va-background-border);
            text-decoration: none;

            &:hover {
              text-decoration: underline;
            }
          }

          &.mono {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.875rem;
            word-break: break-all;
          }
        }
      }
    }
  }

  .providers-card {
    .providers-list {
      display: flex;
      flex-direction: column;
      gap: 1rem;

      .provider-item {
        padding: 1rem;
        background-color: var(--va-background-shade);
        border-radius: 0.5rem;
        border: 1px solid var(--va-background-border);

        .provider-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          gap: 1rem;
          margin-bottom: 1rem;

          h4 {
            margin: 0;
            font-size: 1rem;
            font-weight: 600;
          }
        }

        .provider-details {
          display: flex;
          flex-direction: column;
          gap: 0.75rem;

          .detail-row {
            display: flex;
            align-items: flex-start;
            gap: 1rem;

            label {
              min-width: 100px;
              font-weight: 500;
              font-size: 0.875rem;
              color: var(--va-textColorSecondary);
            }

            span.mono {
              font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
              font-size: 0.875rem;
              word-break: break-all;
            }

            .scopes {
              display: flex;
              flex-wrap: wrap;
              gap: 0.5rem;
            }
          }
        }
      }
    }
  }

  .email-card {
    .email-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
      gap: 1.5rem;

      .email-item {
        label {
          display: block;
          font-size: 0.875rem;
          font-weight: 500;
          color: var(--va-textColorSecondary);
          margin-bottom: 0.5rem;
        }

        p {
          margin: 0;
          font-size: 1rem;
          color: var(--va-textColor);

          &.mono {
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.875rem;
            word-break: break-all;
          }
        }
      }
    }
  }

  .features-card {
    .features-list {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;

      .feature-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 0.75rem;
        background-color: var(--va-background-shade);
        border-radius: 0.5rem;

        .feature-checkbox {
          flex: 1;
        }

        .feature-status {
          font-size: 0.875rem;
          font-weight: 500;
          padding: 0.25rem 0.75rem;
          border-radius: 0.25rem;

          &.enabled {
            color: var(--va-success);
            background-color: rgba(52, 211, 153, 0.1);
          }

          &.disabled {
            color: var(--va-textColorSecondary);
            background-color: rgba(156, 163, 175, 0.1);
          }
        }
      }
    }
  }

  .quotas-card {
    .quotas-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;

      .quota-item {
        padding: 1rem;
        background-color: var(--va-background-shade);
        border-radius: 0.5rem;
        border: 1px solid var(--va-background-border);

        label {
          display: block;
          font-size: 0.875rem;
          font-weight: 500;
          color: var(--va-textColorSecondary);
          margin-bottom: 0.5rem;
        }

        p {
          margin: 0;
          font-size: 1.25rem;
          font-weight: 600;
          color: var(--va-textColor);
        }
      }
    }
  }

  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    justify-content: center;
    align-items: center;
    background-color: rgba(255, 255, 255, 0.5);
    border-radius: 0.5rem;
    z-index: 10;
  }
}

.mr-2 {
  margin-right: 0.5rem;
}

.mono {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style>
