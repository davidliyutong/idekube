<template>
  <va-modal
    v-model="show"
    title="Create Workspace"
    size="large"
    @update:model-value="handleClose"
  >
    <va-form ref="formRef" @submit.prevent="handleSubmit">
      <!-- Basic Info -->
      <va-input
        v-model="form.name"
        label="Workspace Name *"
        placeholder="Enter workspace name"
        :rules="[(v) => !!v || 'Name is required']"
        class="mb-4"
      />

      <va-textarea
        v-model="form.description"
        label="Description"
        placeholder="Enter workspace description"
        :min-rows="2"
        :max-rows="4"
        class="mb-4"
      />

      <!-- Template Selection -->
      <va-select
        v-model="form.template_id"
        label="Template *"
        :options="templateOptions"
        :loading="templateStore.loading"
        :rules="[(v) => !!v || 'Template is required']"
        placeholder="Select a template"
        class="mb-4"
        @update:model-value="handleTemplateChange"
      />

      <div v-if="selectedTemplate" class="template-info mb-4">
        <va-card color="background-element">
          <va-card-content>
            <div class="template-details">
              <div><strong>Image:</strong> <code>{{ selectedTemplate.image }}</code></div>
              <div v-if="selectedTemplate.default_cpu_limit">
                <strong>Default CPU:</strong> {{ selectedTemplate.default_cpu_limit }}
              </div>
              <div v-if="selectedTemplate.default_memory_limit">
                <strong>Default Memory:</strong> {{ formatBytes(selectedTemplate.default_memory_limit) }}
              </div>
            </div>
          </va-card-content>
        </va-card>
      </div>

      <!-- Resource Limits -->
      <div class="section-title">Resource Limits</div>

      <va-input
        v-model.number="form.cpu_limit"
        type="number"
        label="CPU Limit (cores)"
        placeholder="e.g., 2"
        :min="0.5"
        :step="0.5"
        class="mb-4"
      />

      <va-input
        v-model.number="form.memory_limit"
        type="number"
        label="Memory Limit (GB)"
        placeholder="e.g., 4"
        :min="1"
        :step="1"
        class="mb-4"
      />

      <va-input
        v-model.number="form.storage_limit"
        type="number"
        label="Storage Limit (GB)"
        placeholder="e.g., 10"
        :min="1"
        :step="1"
        class="mb-4"
      />

      <!-- Environment Variables -->
      <div class="section-title">Environment Variables</div>

      <div class="env-vars mb-4">
        <div v-for="(env, index) in envVars" :key="index" class="env-var-row">
          <va-input
            v-model="env.key"
            placeholder="Key"
            class="env-key"
          />
          <va-input
            v-model="env.value"
            placeholder="Value"
            class="env-value"
          />
          <va-button
            preset="secondary"
            size="small"
            @click="removeEnvVar(index)"
          >
            <va-icon name="delete" />
          </va-button>
        </div>
        <va-button preset="secondary" size="small" @click="addEnvVar">
          <va-icon name="add" class="mr-2" />
          Add Variable
        </va-button>
      </div>

      <!-- Volume Selection -->
      <div class="section-title">Attach Volumes</div>

      <va-select
        v-model="form.volumes"
        label="Volumes"
        :options="volumeOptions"
        :loading="volumeStore.loading"
        placeholder="Select volumes to attach"
        multiple
        class="mb-4"
      />
    </va-form>

    <template #footer>
      <div class="modal-footer">
        <va-button preset="secondary" @click="handleClose">Cancel</va-button>
        <va-button
          :loading="workspaceStore.loading"
          @click="handleSubmit"
        >
          Create Workspace
        </va-button>
      </div>
    </template>
  </va-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { useWorkspaceStore } from '@/stores/workspace'
import { useTemplateStore } from '@/stores/template'
import { useVolumeStore } from '@/stores/volume'

interface Props {
  modelValue: boolean
  preselectedTemplateId?: string
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
}

const props = defineProps<Props>()
const emit = defineEmits<Emits>()

const workspaceStore = useWorkspaceStore()
const templateStore = useTemplateStore()
const volumeStore = useVolumeStore()

const formRef = ref()
const show = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
})

// Form state
const form = ref({
  name: '',
  description: '',
  template_id: '',
  cpu_limit: undefined as number | undefined,
  memory_limit: undefined as number | undefined,
  storage_limit: undefined as number | undefined,
  volumes: [] as string[],
})

const envVars = ref<Array<{ key: string; value: string }>>([])

// Computed
const templateOptions = computed(() =>
  templateStore.templates.map((t) => ({
    text: t.name,
    value: t.id,
  }))
)

const volumeOptions = computed(() =>
  volumeStore.availableVolumes.map((v) => ({
    text: `${v.name} (${formatBytes(v.size)})`,
    value: v.id,
  }))
)

const selectedTemplate = computed(() => {
  if (!form.value.template_id) return null
  return templateStore.templateById(form.value.template_id)
})

// Methods
function formatBytes(bytes: number): string {
  const gb = bytes / (1024 * 1024 * 1024)
  return `${gb.toFixed(2)} GB`
}

function handleTemplateChange(): void {
  const template = selectedTemplate.value
  if (template) {
    // Set default values from template
    if (template.default_cpu_limit) {
      form.value.cpu_limit = template.default_cpu_limit
    }
    if (template.default_memory_limit) {
      form.value.memory_limit = template.default_memory_limit / (1024 * 1024 * 1024)
    }
    if (template.default_storage_limit) {
      form.value.storage_limit = template.default_storage_limit / (1024 * 1024 * 1024)
    }

    // Load default environment variables
    if (template.default_env_vars) {
      envVars.value = Object.entries(template.default_env_vars).map(([key, value]) => ({
        key,
        value,
      }))
    }
  }
}

function addEnvVar(): void {
  envVars.value.push({ key: '', value: '' })
}

function removeEnvVar(index: number): void {
  envVars.value.splice(index, 1)
}

function handleClose(): void {
  resetForm()
  emit('update:modelValue', false)
}

function resetForm(): void {
  form.value = {
    name: '',
    description: '',
    template_id: '',
    cpu_limit: undefined,
    memory_limit: undefined,
    storage_limit: undefined,
    volumes: [],
  }
  envVars.value = []
}

async function handleSubmit(): Promise<void> {
  const valid = formRef.value?.validate()
  if (!valid) return

  try {
    const env_vars: Record<string, string> = {}
    envVars.value.forEach((env) => {
      if (env.key && env.value) {
        env_vars[env.key] = env.value
      }
    })

    await workspaceStore.createWorkspace({
      name: form.value.name,
      description: form.value.description || undefined,
      template_id: form.value.template_id,
      cpu_limit: form.value.cpu_limit,
      memory_limit: form.value.memory_limit
        ? form.value.memory_limit * 1024 * 1024 * 1024
        : undefined,
      storage_limit: form.value.storage_limit
        ? form.value.storage_limit * 1024 * 1024 * 1024
        : undefined,
      env_vars: Object.keys(env_vars).length > 0 ? env_vars : undefined,
      volumes: form.value.volumes.length > 0 ? form.value.volumes : undefined,
    })

    emit('created')
    handleClose()
  } catch (error) {
    console.error('Failed to create workspace:', error)
  }
}

// Watch for preselected template
watch(
  () => props.preselectedTemplateId,
  (templateId) => {
    if (templateId) {
      form.value.template_id = templateId
      handleTemplateChange()
    }
  },
  { immediate: true }
)

// Lifecycle
onMounted(() => {
  if (templateStore.templates.length === 0) {
    templateStore.fetchTemplates()
  }
  if (volumeStore.volumes.length === 0) {
    volumeStore.fetchVolumes()
  }
})
</script>

<style lang="scss" scoped>
.section-title {
  font-size: 1.1rem;
  font-weight: 600;
  margin-bottom: 1rem;
  margin-top: 1rem;
}

.template-info {
  .template-details {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    font-size: 0.875rem;

    code {
      background-color: var(--va-background-element);
      padding: 0.125rem 0.25rem;
      border-radius: 3px;
      font-size: 0.8rem;
    }
  }
}

.env-vars {
  .env-var-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto;
    gap: 0.5rem;
    margin-bottom: 0.5rem;

    .env-key,
    .env-value {
      margin: 0;
    }
  }
}

.modal-footer {
  display: flex;
  gap: 0.5rem;
  justify-content: flex-end;
}
</style>
