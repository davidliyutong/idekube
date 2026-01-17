<template>
  <va-modal
    v-model="isOpen"
    :title="isEditing ? '编辑组织' : '创建新组织'"
    size="large"
    @ok="handleSubmit"
  >
    <div class="modal-form">
      <va-form ref="form" @submit.prevent="handleSubmit">
        <!-- Organization Name -->
        <div class="form-group">
          <label class="form-label">
            组织名称 <span class="required">*</span>
          </label>
          <va-input
            v-model="formData.name"
            placeholder="输入组织名称"
            :rules="[nameValidation]"
            @blur="form?.validate()"
          />
        </div>

        <!-- Description -->
        <div class="form-group">
          <label class="form-label">描述</label>
          <va-textarea
            v-model="formData.description"
            placeholder="输入组织描述（可选）"
            rows="3"
          />
        </div>

        <!-- Workspace Quota -->
        <div class="form-group">
          <label class="form-label">工作空间配额</label>
          <va-input
            v-model.number="formData.workspace_quota"
            type="number"
            placeholder="输入工作空间数量限制（可选）"
            min="0"
          />
          <small class="text-secondary">组织可以创建的最大工作空间数量，0表示无限制</small>
        </div>

        <!-- Storage Quota -->
        <div class="form-group">
          <label class="form-label">存储配额</label>
          <div class="storage-input-group">
            <va-input
              v-model.number="formData.storage_value"
              type="number"
              placeholder="输入大小"
              :rules="[storageValidation]"
              @blur="form?.validate()"
              min="1"
              max="10000"
            />
            <va-select
              v-model="formData.storage_unit"
              :options="storageUnits"
              class="unit-select"
            />
          </div>
          <small class="text-secondary">
            实际大小: {{ formatSize(calculateStorageBytes()) }}
          </small>
        </div>

        <!-- Status (只在编辑时显示) -->
        <div v-if="isEditing" class="form-group">
          <label class="form-label">组织状态</label>
          <va-select
            v-model="formData.status"
            :options="statusOptions"
            placeholder="选择状态"
          />
        </div>

        <!-- Info Box -->
        <div class="info-box">
          <va-icon name="info" class="mr-2" />
          <div>
            <strong>提示:</strong> 组织创建后可以添加成员、分配工作空间和管理配额。
          </div>
        </div>
      </va-form>
    </div>
  </va-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import {
  useOrganizationStore,
  type Organization,
  type OrganizationCreateData,
  type OrganizationUpdateData,
} from '@/stores/organization'

interface Props {
  modelValue: boolean
  editingOrganization?: Organization | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
  (e: 'updated'): void
}

const props = withDefaults(defineProps<Props>(), {
  editingOrganization: null,
})

const emit = defineEmits<Emits>()

const organizationStore = useOrganizationStore()
const form = ref<any>()

// Form data
const formData = ref({
  name: '',
  description: '',
  workspace_quota: 0,
  storage_value: 100,
  storage_unit: 'GB' as 'B' | 'KB' | 'MB' | 'GB' | 'TB',
  status: 'active' as string,
})

// Storage units
const storageUnits = [
  { text: 'B', value: 'B' },
  { text: 'KB', value: 'KB' },
  { text: 'MB', value: 'MB' },
  { text: 'GB', value: 'GB' },
  { text: 'TB', value: 'TB' },
]

const statusOptions = [
  { text: '活跃', value: 'active' },
  { text: '非活跃', value: 'inactive' },
  { text: '禁用', value: 'suspended' },
]

// Computed
const isOpen = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const isEditing = computed(() => !!props.editingOrganization)

// Validations
const nameValidation = (value: string) => {
  if (!value) return '组织名称是必填的'
  if (value.length < 2) return '组织名称最少2个字符'
  if (value.length > 100) return '组织名称不能超过100个字符'
  return true
}

const storageValidation = (value: string | number) => {
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (!numValue || numValue <= 0) return '存储大小必须大于0'
  if (numValue > 10000) return '存储大小不能超过10000'
  return true
}

// Methods
function calculateStorageBytes(): number {
  const units: Record<string, number> = {
    B: 1,
    KB: 1024,
    MB: 1024 * 1024,
    GB: 1024 * 1024 * 1024,
    TB: 1024 * 1024 * 1024 * 1024,
  }
  return formData.value.storage_value * (units[formData.value.storage_unit] || 1)
}

function formatSize(bytes: number): string {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let unitIndex = 0

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024
    unitIndex++
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`
}

async function handleSubmit() {
  if (!form.value) return

  const isValid = await form.value.validate()
  if (!isValid) return

  try {
    if (isEditing.value && props.editingOrganization) {
      // Update existing organization
      const updateData: OrganizationUpdateData = {
        name: formData.value.name,
        description: formData.value.description,
        workspace_quota: formData.value.workspace_quota || undefined,
        storage_quota: calculateStorageBytes() || undefined,
        status: formData.value.status,
      }
      await organizationStore.updateOrganization(props.editingOrganization.id, updateData)
      emit('updated')
    } else {
      // Create new organization
      const createData: OrganizationCreateData = {
        name: formData.value.name,
        description: formData.value.description,
        workspace_quota: formData.value.workspace_quota || undefined,
        storage_quota: calculateStorageBytes() || undefined,
      }
      await organizationStore.createOrganization(createData)
      emit('created')
    }
  } catch (error) {
    console.error('Failed to save organization:', error)
  }
}

function resetForm() {
  formData.value = {
    name: '',
    description: '',
    workspace_quota: 0,
    storage_value: 100,
    storage_unit: 'GB',
    status: 'active',
  }
}

// Watch for editing organization changes
watch(
  () => props.editingOrganization,
  (organization) => {
    if (organization) {
      formData.value.name = organization.name
      formData.value.description = organization.description || ''
      formData.value.workspace_quota = organization.workspace_quota || 0
      formData.value.status = organization.status

      // Parse storage to GB
      const gb = (organization.storage_quota || 0) / (1024 * 1024 * 1024)
      formData.value.storage_value = Math.round(gb * 100) / 100
      formData.value.storage_unit = 'GB'
    } else {
      resetForm()
    }
  }
)

// Watch for modal open/close
watch(isOpen, (newVal) => {
  if (!newVal) {
    resetForm()
  }
})
</script>

<style lang="scss" scoped>
.modal-form {
  padding: 1rem;

  .form-group {
    margin-bottom: 1.5rem;

    &:last-child {
      margin-bottom: 0;
    }

    .form-label {
      display: block;
      margin-bottom: 0.5rem;
      font-weight: 500;

      .required {
        color: var(--va-danger);
      }
    }

    .text-secondary {
      display: block;
      margin-top: 0.25rem;
      color: var(--va-textColorSecondary);
      font-size: 0.875rem;
    }

    .storage-input-group {
      display: flex;
      gap: 0.5rem;
      align-items: center;

      input {
        flex: 1;
      }

      .unit-select {
        min-width: 100px;
      }
    }
  }

  .info-box {
    display: flex;
    gap: 0.75rem;
    padding: 1rem;
    background-color: var(--va-background-shade);
    border-radius: 0.5rem;
    border-left: 4px solid var(--va-info);
    margin-top: 1.5rem;
    font-size: 0.875rem;

    div {
      flex: 1;
    }
  }
}

.mr-2 {
  margin-right: 0.5rem;
}
</style>
