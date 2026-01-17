<template>
  <va-modal
    v-model="isOpen"
    :title="isEditing ? '编辑数据卷' : '创建新数据卷'"
    size="large"
    @ok="handleSubmit"
  >
    <div class="modal-form">
      <va-form ref="form" @submit.prevent="handleSubmit">
        <!-- Name -->
        <div class="form-group">
          <label class="form-label">
            数据卷名称 <span class="required">*</span>
          </label>
          <va-input
            v-model="formData.name"
            placeholder="输入数据卷名称"
            :rules="[nameValidation]"
            @blur="form?.validate()"
          />
        </div>

        <!-- Description -->
        <div class="form-group">
          <label class="form-label">描述</label>
          <va-textarea
            v-model="formData.description"
            placeholder="输入数据卷描述（可选）"
            rows="3"
          />
        </div>

        <!-- Size -->
        <div class="form-group">
          <label class="form-label">
            大小 <span class="required">*</span>
          </label>
          <div class="size-input-group">
            <va-input
              v-model.number="formData.sizeValue"
              type="number"
              placeholder="输入大小"
              :rules="[sizeValidation]"
              @blur="form?.validate()"
              min="1"
              max="1000"
            />
            <va-select
              v-model="formData.sizeUnit"
              :options="sizeUnits"
              class="unit-select"
            />
          </div>
          <small class="text-secondary">
            实际大小: {{ formatSize(calculateBytes()) }}
          </small>
        </div>

        <!-- Organization (if applicable) -->
        <div v-if="availableOrganizations.length > 0" class="form-group">
          <label class="form-label">组织 (可选)</label>
          <va-select
            v-model="formData.organization_id"
            :options="organizationOptions"
            placeholder="选择组织"
            clearable
          />
        </div>

        <!-- Labels -->
        <div class="form-group">
          <label class="form-label">标签 (可选)</label>
          <div class="labels-input">
            <div v-for="(_, key) in formData.labels" :key="key" class="label-pair">
              <va-input
                v-model="formData.labels[key]"
                placeholder="值"
                class="label-value-input"
              />
              <va-button
                size="small"
                color="danger"
                preset="secondary"
                @click="removeLabel(key)"
              >
                <va-icon name="delete" />
              </va-button>
            </div>
            <va-button size="small" preset="secondary" @click="addLabel">
              <va-icon name="add" class="mr-1" />
              添加标签
            </va-button>
          </div>
        </div>

        <!-- Info Box -->
        <div class="info-box">
          <va-icon name="info" class="mr-2" />
          <div>
            <strong>提示:</strong> 数据卷创建后，您可以将其挂载到工作空间中使用。
          </div>
        </div>
      </va-form>
    </div>
  </va-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useVolumeStore, type Volume, type VolumeCreateData, type VolumeUpdateData } from '@/stores/volume'

interface Props {
  modelValue: boolean
  editingVolume?: Volume | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
  (e: 'updated'): void
}

const props = withDefaults(defineProps<Props>(), {
  editingVolume: null,
})

const emit = defineEmits<Emits>()

const volumeStore = useVolumeStore()
const form = ref<any>()

// Form data
const formData = ref({
  name: '',
  description: '',
  sizeValue: 10,
  sizeUnit: 'GB' as 'B' | 'KB' | 'MB' | 'GB' | 'TB',
  organization_id: '',
  labels: {} as Record<string, string>,
})

// Size units
const sizeUnits = [
  { text: 'B', value: 'B' },
  { text: 'KB', value: 'KB' },
  { text: 'MB', value: 'MB' },
  { text: 'GB', value: 'GB' },
  { text: 'TB', value: 'TB' },
]

// Available organizations (can be fetched from organization store)
const availableOrganizations = ref<any[]>([])

const organizationOptions = computed(() =>
  availableOrganizations.value.map((org) => ({
    text: org.name,
    value: org.id,
  }))
)

// Computed
const isOpen = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const isEditing = computed(() => !!props.editingVolume)

// Validations
const nameValidation = (value: string) => {
  if (!value) return 'Data volume name is required'
  if (value.length < 3) return 'Name must be at least 3 characters'
  if (value.length > 255) return 'Name must not exceed 255 characters'
  return true
}

const sizeValidation = (value: string | number) => {
  const numValue = typeof value === 'string' ? parseFloat(value) : value
  if (!numValue || numValue <= 0) return 'Size must be greater than 0'
  if (numValue > 1000) return 'Size must not exceed 1000'
  return true
}

// Methods
function calculateBytes(): number {
  const units: Record<string, number> = {
    B: 1,
    KB: 1024,
    MB: 1024 * 1024,
    GB: 1024 * 1024 * 1024,
    TB: 1024 * 1024 * 1024 * 1024,
  }
  return formData.value.sizeValue * (units[formData.value.sizeUnit] || 1)
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

function addLabel() {
  const key = `label_${Date.now()}`
  formData.value.labels[key] = ''
}

function removeLabel(key: string) {
  delete formData.value.labels[key]
}

async function handleSubmit() {
  if (!form.value) return

  const isValid = await form.value.validate()
  if (!isValid) return

  try {
    if (isEditing.value && props.editingVolume) {
      // Update existing volume
      const updateData: VolumeUpdateData = {
        name: formData.value.name,
        description: formData.value.description,
        size: calculateBytes(),
        labels: Object.keys(formData.value.labels).length > 0 ? formData.value.labels : undefined,
      }
      await volumeStore.updateVolume(props.editingVolume.id, updateData)
      emit('updated')
    } else {
      // Create new volume
      const createData: VolumeCreateData = {
        name: formData.value.name,
        description: formData.value.description,
        size: calculateBytes(),
        organization_id: formData.value.organization_id || undefined,
        labels: Object.keys(formData.value.labels).length > 0 ? formData.value.labels : undefined,
      }
      await volumeStore.createVolume(createData)
      emit('created')
    }
  } catch (error) {
    console.error('Failed to save volume:', error)
  }
}

function resetForm() {
  formData.value = {
    name: '',
    description: '',
    sizeValue: 10,
    sizeUnit: 'GB',
    organization_id: '',
    labels: {},
  }
}

// Watch for editing volume changes
watch(
  () => props.editingVolume,
  (volume) => {
    if (volume) {
      formData.value.name = volume.name
      formData.value.description = volume.description || ''
      formData.value.organization_id = volume.organization_id || ''
      formData.value.labels = { ...volume.labels } || {}

      // Parse size to GB
      const gb = volume.size / (1024 * 1024 * 1024)
      formData.value.sizeValue = Math.round(gb * 100) / 100
      formData.value.sizeUnit = 'GB'
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

    .size-input-group {
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

    .text-secondary {
      display: block;
      margin-top: 0.25rem;
      color: var(--va-textColorSecondary);
      font-size: 0.875rem;
    }

    .labels-input {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;

      .label-pair {
        display: flex;
        gap: 0.5rem;
        align-items: center;

        .label-value-input {
          flex: 1;
        }
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

.mr-1 {
  margin-right: 0.25rem;
}

.mr-2 {
  margin-right: 0.5rem;
}
</style>
