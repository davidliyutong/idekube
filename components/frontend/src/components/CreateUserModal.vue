<template>
  <va-modal
    v-model="isOpen"
    :title="isEditing ? '编辑用户' : '创建新用户'"
    size="large"
    @ok="handleSubmit"
  >
    <div class="modal-form">
      <va-form ref="form" @submit.prevent="handleSubmit">
        <!-- Username -->
        <div class="form-group">
          <label class="form-label">
            用户名 <span class="required">*</span>
          </label>
          <va-input
            v-model="formData.username"
            placeholder="输入用户名"
            :disabled="isEditing"
            :rules="[usernameValidation]"
            @blur="form?.validate()"
          />
          <small class="text-secondary" v-if="!isEditing">最少3个字符，只能含字母、数字、下划线</small>
        </div>

        <!-- Email -->
        <div class="form-group">
          <label class="form-label">
            邮箱 <span class="required">*</span>
          </label>
          <va-input
            v-model="formData.email"
            type="email"
            placeholder="输入邮箱地址"
            :rules="[emailValidation]"
            @blur="form?.validate()"
          />
        </div>

        <!-- Full Name -->
        <div class="form-group">
          <label class="form-label">全名</label>
          <va-input
            v-model="formData.full_name"
            placeholder="输入用户全名（可选）"
          />
        </div>

        <!-- Password (只在创建时显示) -->
        <div v-if="!isEditing" class="form-group">
          <label class="form-label">
            密码 <span class="required">*</span>
          </label>
          <va-input
            v-model="formData.password"
            type="password"
            placeholder="设置用户密码"
            :rules="[passwordValidation]"
            @blur="form?.validate()"
          />
          <small class="text-secondary">最少8个字符，包含大小写字母、数字和特殊字符</small>
        </div>

        <!-- Roles -->
        <div class="form-group">
          <label class="form-label">角色分配</label>
          <div class="roles-grid">
            <va-checkbox
              v-for="role in availableRoles"
              :key="role"
              v-model="formData.roles"
              :option="role"
              class="role-checkbox"
            >
              {{ getRoleLabel(role) }}
            </va-checkbox>
          </div>
        </div>

        <!-- Status (只在编辑时显示) -->
        <div v-if="isEditing" class="form-group">
          <label class="form-label">账户状态</label>
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
            <strong>提示:</strong> 用户创建后可以自行修改密码。管理员可以重置用户密码。
          </div>
        </div>
      </va-form>
    </div>
  </va-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useUserStore, type User, type UserCreateData, type UserUpdateData } from '@/stores/user'

interface Props {
  modelValue: boolean
  editingUser?: User | null
}

interface Emits {
  (e: 'update:modelValue', value: boolean): void
  (e: 'created'): void
  (e: 'updated'): void
}

const props = withDefaults(defineProps<Props>(), {
  editingUser: null,
})

const emit = defineEmits<Emits>()

const userStore = useUserStore()
const form = ref<any>()

// Form data
const formData = ref({
  username: '',
  email: '',
  full_name: '',
  password: '',
  roles: [] as string[],
  status: 'active' as string,
})

const statusOptions = [
  { text: '活跃', value: 'active' },
  { text: '非活跃', value: 'inactive' },
  { text: '禁用', value: 'suspended' },
]

// Available roles
const availableRoles = computed(() => userStore.availableRoles)

// Computed
const isOpen = computed({
  get: () => props.modelValue,
  set: (value: boolean) => emit('update:modelValue', value),
})

const isEditing = computed(() => !!props.editingUser)

// Validations
const usernameValidation = (value: string) => {
  if (!value) return '用户名是必填的'
  if (value.length < 3) return '用户名最少3个字符'
  if (!/^[a-zA-Z0-9_]+$/.test(value)) return '用户名只能包含字母、数字和下划线'
  return true
}

const emailValidation = (value: string) => {
  if (!value) return '邮箱是必填的'
  if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) return '请输入有效的邮箱地址'
  return true
}

const passwordValidation = (value: string) => {
  if (!value && !isEditing.value) return '密码是必填的'
  if (value && value.length < 8) return '密码最少8个字符'
  if (value && !/[A-Z]/.test(value)) return '密码必须包含大写字母'
  if (value && !/[a-z]/.test(value)) return '密码必须包含小写字母'
  if (value && !/[0-9]/.test(value)) return '密码必须包含数字'
  if (value && !/[!@#$%^&*(),.?":{}|<>]/.test(value)) return '密码必须包含特殊字符'
  return true
}

// Methods
function getRoleLabel(role: string): string {
  const roleMap: Record<string, string> = {
    admin: '管理员',
    operator: '运维人员',
    developer: '开发者',
    viewer: '浏览者',
  }
  return roleMap[role] || role
}

async function handleSubmit() {
  if (!form.value) return

  const isValid = await form.value.validate()
  if (!isValid) return

  try {
    if (isEditing.value && props.editingUser) {
      // Update existing user
      const updateData: UserUpdateData = {
        email: formData.value.email,
        full_name: formData.value.full_name,
        roles: formData.value.roles,
        status: formData.value.status,
      }
      await userStore.updateUser(props.editingUser.id, updateData)
      emit('updated')
    } else {
      // Create new user
      const createData: UserCreateData = {
        username: formData.value.username,
        email: formData.value.email,
        full_name: formData.value.full_name,
        password: formData.value.password,
        roles: formData.value.roles,
      }
      await userStore.createUser(createData)
      emit('created')
    }
  } catch (error) {
    console.error('Failed to save user:', error)
  }
}

function resetForm() {
  formData.value = {
    username: '',
    email: '',
    full_name: '',
    password: '',
    roles: [],
    status: 'active',
  }
}

// Watch for editing user changes
watch(
  () => props.editingUser,
  (user) => {
    if (user) {
      formData.value.username = user.username
      formData.value.email = user.email
      formData.value.full_name = user.full_name || ''
      formData.value.roles = [...user.roles]
      formData.value.status = user.status
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

    .roles-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 1rem;

      .role-checkbox {
        padding: 0.5rem;
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
