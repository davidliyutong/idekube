<template>
  <div class="settings-page">
    <div class="page-header">
      <h1 class="page-title">系统设置</h1>
      <div class="header-actions">
        <va-button @click="handleResetForm" preset="secondary" :disabled="loading">
          <va-icon name="refresh" class="mr-2" />
          重置
        </va-button>
        <va-button @click="handleSaveSettings" :disabled="loading || !hasChanges">
          <va-icon name="save" class="mr-2" />
          保存
        </va-button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading && settings.length === 0" class="loading-state">
      <va-progress-circle indeterminate />
    </div>

    <!-- Settings Tabs -->
    <div v-else class="settings-container">
      <va-tabs v-model="activeTab" class="settings-tabs">
        <template v-for="(_, category) in settingsByCategory" :key="category">
          <va-tab :name="category" class="category-tab">
            <template #title>
              {{ getCategoryLabel(category) }}
              <span class="tab-count">({{ settingsByCategory[category].length }})</span>
            </template>

            <div class="tab-content">
              <div
                v-for="setting in settingsByCategory[category]"
                :key="setting.key"
                class="setting-item"
                :class="{ 'setting-public': setting.is_public }"
              >
                <div class="setting-header">
                  <div class="setting-title">
                    <h3>{{ setting.key }}</h3>
                    <p class="setting-description">{{ setting.description }}</p>
                  </div>
                  <div class="setting-badges">
                    <va-chip v-if="setting.is_public" color="info" size="small">
                      Public
                    </va-chip>
                    <va-chip color="secondary" size="small">
                      {{ setting.value_type }}
                    </va-chip>
                  </div>
                </div>

                <div class="setting-input">
                  <va-textarea
                    v-model="editedSettings[setting.key]"
                    :disabled="loading"
                    :rows="getSetting(setting.key)?.value.split('\n').length || 3"
                    class="setting-textarea"
                  />
                  <div class="setting-meta">
                    <span class="updated-time">
                      Updated: {{ formatDate(setting.updated_at) }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </va-tab>
        </template>
      </va-tabs>
    </div>

    <!-- Empty State -->
    <div v-if="!loading && settings.length === 0" class="empty-state">
      <p>No settings found</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useSettingsStore, SETTING_CATEGORIES } from '@/stores/settings'
import { formatDate } from '@/utils/date'

const settingsStore = useSettingsStore()

// Refs
const activeTab = ref(0)
const editedSettings = ref<Record<string, string>>({})
const originalSettings = ref<Record<string, string>>({})

// Computed
const settings = computed(() => settingsStore.settings)
const { loading } = settingsStore
const { settingsByCategory, getSetting } = settingsStore

const hasChanges = computed(() => {
  return Object.keys(editedSettings.value).some(
    (key) => editedSettings.value[key] !== originalSettings.value[key]
  )
})

// Methods
function getCategoryLabel(category: string): string {
  return SETTING_CATEGORIES[category as keyof typeof SETTING_CATEGORIES] || category
}

function initializeEditedSettings() {
  editedSettings.value = {}
  originalSettings.value = {}

  settings.value.forEach((setting) => {
    editedSettings.value[setting.key] = setting.value
    originalSettings.value[setting.key] = setting.value
  })
}

async function handleSaveSettings() {
  const changedSettings: Record<string, string> = {}

  Object.keys(editedSettings.value).forEach((key) => {
    if (editedSettings.value[key] !== originalSettings.value[key]) {
      changedSettings[key] = editedSettings.value[key]
    }
  })

  if (Object.keys(changedSettings).length === 0) {
    return
  }

  try {
    await settingsStore.batchUpdateSettings(changedSettings)
    initializeEditedSettings()
  } catch (error) {
    console.error('Failed to save settings:', error)
  }
}

function handleResetForm() {
  initializeEditedSettings()
}

async function loadSettings() {
  try {
    await settingsStore.fetchSettings()
    initializeEditedSettings()
  } catch (error) {
    console.error('Failed to load settings:', error)
  }
}

// Lifecycle
onMounted(() => {
  loadSettings()
})

// Watch for settings updates
watch(
  () => settingsStore.settings,
  () => {
    initializeEditedSettings()
  }
)
</script>

<style lang="scss" scoped>
.settings-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 1px solid var(--va-background-border);

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }

    .header-actions {
      display: flex;
      gap: 0.5rem;
    }
  }

  .loading-state {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 500px;
  }

  .settings-container {
    .settings-tabs {
      :deep(.va-tabs__content) {
        padding: 0;
      }
    }

    .category-tab {
      .tab-count {
        font-size: 0.875rem;
        color: var(--va-textColorSecondary);
        margin-left: 0.5rem;
      }
    }

    .tab-content {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
      padding: 1.5rem 0;

      .setting-item {
        padding: 1.5rem;
        background-color: var(--va-background-shade);
        border: 1px solid var(--va-background-border);
        border-radius: 0.5rem;
        border-left: 4px solid var(--va-primary);

        &.setting-public {
          border-left-color: var(--va-info);
        }

        .setting-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 1rem;
          gap: 1rem;

          .setting-title {
            flex: 1;

            h3 {
              margin: 0 0 0.5rem 0;
              font-size: 1.125rem;
              font-weight: 600;
              color: var(--va-textColor);
            }

            .setting-description {
              margin: 0;
              font-size: 0.875rem;
              color: var(--va-textColorSecondary);
            }
          }

          .setting-badges {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
            justify-content: flex-end;
          }
        }

        .setting-input {
          .setting-textarea {
            width: 100%;
            min-height: 60px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.875rem;
          }

          .setting-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-top: 0.5rem;
            font-size: 0.75rem;
            color: var(--va-textColorSecondary);
          }
        }
      }
    }
  }

  .empty-state {
    text-align: center;
    padding: 3rem 1rem;
    color: var(--va-textColorSecondary);
    background-color: var(--va-background-shade);
    border-radius: 0.5rem;
    border: 1px dashed var(--va-background-border);
  }
}

.mr-2 {
  margin-right: 0.5rem;
}
</style>
