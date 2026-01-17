<template>
  <div class="permission-checker-page">
    <div class="page-header">
      <h1 class="page-title">Permission Checker</h1>
    </div>

    <va-card>
      <va-card-content>
        <div class="checker-content">
          <!-- Left Panel: Form -->
          <div class="form-panel">
            <h3 class="panel-title">Check Permission</h3>
            
            <div class="form-fields">
              <va-input
                v-model="formData.user_id"
                label="User ID"
                placeholder="Enter user ID"
              />

              <va-select
                v-model="formData.resource_type"
                label="Resource Type"
                :options="resourceTypes"
                placeholder="Select resource type"
              />

              <va-input
                v-model="formData.resource_id"
                label="Resource ID"
                placeholder="Enter resource ID (optional)"
              />

              <va-select
                v-model="formData.action"
                label="Action"
                :options="actions"
                placeholder="Select action"
              />

              <va-button
                @click="checkPermission"
                :loading="loading"
                class="check-btn"
              >
                <va-icon name="security" class="mr-2" />
                Check Permission
              </va-button>
            </div>
          </div>

          <!-- Right Panel: Result -->
          <div class="result-panel">
            <h3 class="panel-title">Result</h3>
            
            <div v-if="!result" class="empty-state">
              <va-icon name="info" size="large" color="secondary" />
              <p>Fill in the form and click "Check Permission" to see the result</p>
            </div>

            <div v-else class="result-content">
              <div class="result-status" :class="{ allowed: result.allowed, denied: !result.allowed }">
                <va-icon 
                  :name="result.allowed ? 'check_circle' : 'cancel'" 
                  size="large" 
                  :color="result.allowed ? 'success' : 'danger'"
                />
                <h2>{{ result.allowed ? 'ALLOWED' : 'DENIED' }}</h2>
              </div>

              <div class="result-details">
                <div class="detail-row">
                  <span class="label">User:</span>
                  <va-chip size="small">{{ lastCheck.user_id }}</va-chip>
                </div>
                <div class="detail-row">
                  <span class="label">Resource:</span>
                  <code>{{ lastCheck.resource_type }}{{ lastCheck.resource_id ? ':' + lastCheck.resource_id : '' }}</code>
                </div>
                <div class="detail-row">
                  <span class="label">Action:</span>
                  <va-chip size="small" color="primary">{{ lastCheck.action }}</va-chip>
                </div>
                <div class="detail-row">
                  <span class="label">Result:</span>
                  <va-badge 
                    :color="result.allowed ? 'success' : 'danger'"
                    :text="result.allowed ? 'Access Granted' : 'Access Denied'"
                  />
                </div>
              </div>

              <va-alert 
                v-if="!result.allowed" 
                color="warning"
                class="mt-4"
              >
                <template #title>Access Denied</template>
                The user does not have permission to perform this action on the specified resource.
              </va-alert>
            </div>
          </div>
        </div>
      </va-card-content>
    </va-card>

    <!-- Recent Checks -->
    <va-card v-if="recentChecks.length > 0" class="mt-4">
      <va-card-title>Recent Checks</va-card-title>
      <va-card-content>
        <va-data-table
          :items="recentChecks"
          :columns="historyColumns"
          striped
        >
          <template #cell(resource)="{ row }">
            <code>{{ row.resource_type }}{{ row.resource_id ? ':' + row.resource_id : '' }}</code>
          </template>

          <template #cell(action)="{ row }">
            <va-chip size="small" color="primary">{{ row.action }}</va-chip>
          </template>

          <template #cell(result)="{ row }">
            <va-badge 
              :color="row.allowed ? 'success' : 'danger'"
              :text="row.allowed ? 'Allowed' : 'Denied'"
            />
          </template>

          <template #cell(timestamp)="{ row }">
            <span>{{ formatDate(row.timestamp) }}</span>
          </template>
        </va-data-table>
      </va-card-content>
    </va-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import apiClient from '@/api/client'
import { formatDate } from '@/utils/date'
import { useNotification } from '@/composables/useNotification'

interface PermissionCheckRequest {
  user_id: string
  resource_type: string
  resource_id?: string
  action: string
}

interface PermissionCheckResult {
  allowed: boolean
}

interface CheckHistory extends PermissionCheckRequest {
  allowed: boolean
  timestamp: Date
}

const { error: showError } = useNotification()

const loading = ref(false)
const result = ref<PermissionCheckResult | null>(null)
const recentChecks = ref<CheckHistory[]>([])

const formData = ref<PermissionCheckRequest>({
  user_id: '',
  resource_type: '',
  resource_id: '',
  action: '',
})

const lastCheck = ref<PermissionCheckRequest>({
  user_id: '',
  resource_type: '',
  resource_id: '',
  action: '',
})

const resourceTypes = [
  'workspace',
  'template',
  'volume',
  'user',
  'organization',
  'api-key',
  'webhook',
  'policy',
]

const actions = [
  'read',
  'write',
  'delete',
  'create',
  'update',
  'start',
  'stop',
  'restart',
  'manage',
]

const historyColumns = [
  { key: 'user_id', label: 'User ID' },
  { key: 'resource', label: 'Resource' },
  { key: 'action', label: 'Action' },
  { key: 'result', label: 'Result' },
  { key: 'timestamp', label: 'Time' },
]

async function checkPermission() {
  if (!formData.value.user_id || !formData.value.resource_type || !formData.value.action) {
    showError('Please fill in all required fields')
    return
  }

  loading.value = true
  try {
    const data: any = {
      user_id: formData.value.user_id,
      resource_type: formData.value.resource_type,
      action: formData.value.action,
    }
    if (formData.value.resource_id) {
      data.resource_id = formData.value.resource_id
    }

    const response = await apiClient.post<PermissionCheckResult>('/permissions/check', data)
    result.value = response.data
    lastCheck.value = { ...formData.value }

    // Add to recent checks
    recentChecks.value.unshift({
      ...formData.value,
      allowed: response.data.allowed,
      timestamp: new Date(),
    })

    // Keep only last 10 checks
    if (recentChecks.value.length > 10) {
      recentChecks.value = recentChecks.value.slice(0, 10)
    }
  } catch (err: any) {
    showError('Failed to check permission')
    console.error('Permission check error:', err)
  } finally {
    loading.value = false
  }
}
</script>

<style lang="scss" scoped>
.permission-checker-page {
  .page-header {
    margin-bottom: 2rem;

    .page-title {
      margin: 0;
      font-size: 2rem;
      font-weight: 600;
    }
  }

  .checker-content {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2rem;

    @media (max-width: 768px) {
      grid-template-columns: 1fr;
    }

    .panel-title {
      margin: 0 0 1.5rem 0;
      font-size: 1.25rem;
      font-weight: 600;
    }

    .form-panel {
      .form-fields {
        display: flex;
        flex-direction: column;
        gap: 1rem;

        .check-btn {
          margin-top: 0.5rem;
        }
      }
    }

    .result-panel {
      .empty-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        padding: 3rem 1rem;
        text-align: center;
        color: var(--va-textColorSecondary);

        p {
          margin-top: 1rem;
          max-width: 300px;
        }
      }

      .result-content {
        .result-status {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 2rem;
          border-radius: 0.5rem;
          margin-bottom: 1.5rem;

          &.allowed {
            background-color: rgba(76, 175, 80, 0.1);
          }

          &.denied {
            background-color: rgba(244, 67, 54, 0.1);
          }

          h2 {
            margin: 0.5rem 0 0 0;
            font-size: 1.5rem;
            font-weight: 600;
          }
        }

        .result-details {
          display: flex;
          flex-direction: column;
          gap: 1rem;

          .detail-row {
            display: flex;
            align-items: center;
            gap: 0.75rem;

            .label {
              min-width: 80px;
              font-weight: 500;
              color: var(--va-textColorSecondary);
            }

            code {
              font-family: 'Monaco', 'Menlo', monospace;
              font-size: 0.875rem;
              background-color: var(--va-background-shade);
              padding: 0.25rem 0.5rem;
              border-radius: 0.25rem;
            }
          }
        }
      }
    }
  }
}

.mr-2 { margin-right: 0.5rem; }
.mt-4 { margin-top: 1rem; }
</style>
