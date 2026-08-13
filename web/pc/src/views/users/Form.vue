<template>
  <div class="user-form">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ isEdit ? '编辑用户' : '创建用户' }}</span>
          <el-button @click="handleBack">返回</el-button>
        </div>
      </template>

      <el-form
        ref="formRef"
        :model="formData"
        :rules="rules"
        label-width="120px"
        style="max-width: 600px"
      >
        <el-form-item label="手机号" prop="phone">
          <el-input
            v-model="formData.phone"
            placeholder="请输入11位手机号"
            maxlength="11"
            :disabled="isEdit"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input
            v-model="formData.password"
            type="password"
            placeholder="请输入密码（6-20位）"
            show-password
          />
        </el-form-item>

        <el-form-item label="昵称" prop="nickname">
          <el-input
            v-model="formData.nickname"
            placeholder="请输入昵称"
            maxlength="50"
          />
        </el-form-item>

        <el-form-item label="状态" prop="status" v-if="isEdit">
          <el-select v-model="formData.status" placeholder="请选择状态">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>

        <el-form-item label="创建时间" v-if="isEdit && formData.created_at">
          <span>{{ formatTime(formData.created_at) }}</span>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
          <el-button @click="handleBack">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
/* eslint-disable */
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage, type FormInstance, type FormRules } from 'element-plus'
import { getUserById, createUser, updateUser } from '@/api/identity'

const router = useRouter()
const route = useRoute()
const formRef = ref<FormInstance>()
const submitting = ref(false)

const userId = computed(() => {
  const id = route.params.id as string | undefined
  return id || null
})

const isEdit = computed(() => userId.value !== null)

const formData = reactive({
  phone: '',
  password: '',
  nickname: '',
  status: 1,
  created_at: '' as string | number
})

// Phone validation
const validatePhone = (_rule: any, value: string, callback: any) => {
  if (!value) {
    callback(new Error('请输入手机号'))
  } else if (!/^1[3-9]\d{9}$/.test(value)) {
    callback(new Error('请输入正确的手机号'))
  } else {
    callback()
  }
}

// Password validation
const validatePassword = (_rule: any, value: string, callback: any) => {
  if (!isEdit.value && !value) {
    callback(new Error('请输入密码'))
  } else if (value && (value.length < 6 || value.length > 20)) {
    callback(new Error('密码长度为6-20位'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  phone: [{ validator: validatePhone, trigger: 'blur' }],
  password: [{ validator: validatePassword, trigger: 'blur' }],
  nickname: [
    { required: true, message: '请输入昵称', trigger: 'blur' },
    { min: 2, max: 50, message: '昵称长度为2-50个字符', trigger: 'blur' }
  ],
}

const loadUser = async () => {
  if (!userId.value) return

  try {
    const user = await getUserById(userId.value)
    formData.phone = user.phone
    formData.nickname = user.nickname || ''
    formData.status = user.status
    formData.created_at = user.created_at
  } catch (error) {
    ElMessage.error('加载用户信息失败')
    console.error('Load user error:', error)
    handleBack()
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    submitting.value = true

    if (isEdit.value) {
      await updateUser(userId.value!, {
        nickname: formData.nickname,
        status: formData.status
      })
      ElMessage.success('用户更新成功')
    } else {
      await createUser({
        phone: formData.phone,
        password: formData.password,
        nickname: formData.nickname,
        user_type: 2
      })
      ElMessage.success('用户创建成功')
    }

    handleBack()
  } catch (error: any) {
    if (error !== false) {
      const msg = error?.message || ''
      if (msg.includes('已注册') || msg.includes('10002')) {
        ElMessage.error('该手机号已注册，请更换')
      } else {
        ElMessage.error(isEdit.value ? '用户更新失败' : '用户创建失败')
      }
      console.error('Submit user error:', error)
    }
  } finally {
    submitting.value = false
  }
}

const formatTime = (ts: string | number) => {
  if (!ts) return '-'
  const d = new Date(Number(ts) * 1000)
  return d.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

const handleBack = () => {
  router.push({ name: 'UserList' })
}

onMounted(() => {
  if (isEdit.value) {
    loadUser()
  }
})
</script>

<style scoped>
.user-form {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}
</style>
