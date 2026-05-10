<template>
  <div class="register-container">
    <div class="register-box">
      <div class="register-header">
        <h1>用户注册</h1>
        <p>Community & Home Management System</p>
      </div>

      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        class="register-form"
        @submit.prevent="handleRegister"
      >
        <el-form-item prop="phone">
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            prefix-icon="Phone"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="smsCode">
          <div class="sms-input-group">
            <el-input
              v-model="form.smsCode"
              placeholder="请输入验证码"
              prefix-icon="Message"
              size="large"
              maxlength="6"
            />
            <el-button
              :disabled="smsCountdown > 0"
              size="large"
              @click="handleSendSms"
            >
              {{ smsCountdown > 0 ? `${smsCountdown}秒后重试` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <el-form-item prop="nickname">
          <el-input
            v-model="form.nickname"
            placeholder="请输入昵称（2-20个字符）"
            prefix-icon="User"
            size="large"
          />
        </el-form-item>

        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码（至少8位，包含大小写字母、数字和特殊字符）"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>

        <el-form-item prop="confirmPassword">
          <el-input
            v-model="form.confirmPassword"
            type="password"
            placeholder="请确认密码"
            prefix-icon="Lock"
            size="large"
            show-password
          />
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            style="width: 100%"
            @click="handleRegister"
          >
            注册
          </el-button>
        </el-form-item>
      </el-form>

      <div class="register-footer">
        <router-link to="/login">已有账号？立即登录</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter } from 'vue-router';
import { ElMessage, type FormInstance, type FormRules } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import { sendSms } from '@/api/identity';
import { requiredPhoneRule, nicknameRule, smsCodeRule } from '@/utils/validation';

const router = useRouter();
const authStore = useAuthStore();

const formRef = ref<FormInstance>();
const loading = ref(false);
const smsCountdown = ref(0);

const form = reactive({
  phone: '',
  smsCode: '',
  nickname: '',
  password: '',
  confirmPassword: ''
});

// Custom validator for password confirmation
const validateConfirmPassword = (_rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请确认密码'));
  } else if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'));
  } else {
    callback();
  }
};

const rules: FormRules = {
  phone: requiredPhoneRule,
  smsCode: smsCodeRule,
  nickname: nicknameRule,
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, message: '密码至少8位', trigger: 'blur' },
    {
      pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
      message: '密码必须包含大小写字母、数字和特殊字符',
      trigger: 'blur'
    }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirmPassword, trigger: 'blur' }
  ]
};

// Send SMS code
const handleSendSms = async () => {
  if (!form.phone) {
    ElMessage.warning('请输入手机号');
    return;
  }

  if (!/^1[3-9]\d{9}$/.test(form.phone)) {
    ElMessage.warning('请输入正确的手机号');
    return;
  }

  try {
    await sendSms(form.phone);
    ElMessage.success('验证码已发送');

    // Start countdown
    smsCountdown.value = 60;
    const timer = setInterval(() => {
      smsCountdown.value--;
      if (smsCountdown.value <= 0) {
        clearInterval(timer);
      }
    }, 1000);
  } catch (error: any) {
    ElMessage.error(error.message || '发送失败');
  }
};

// Handle registration
const handleRegister = async () => {
  if (!formRef.value) return;

  await formRef.value.validate(async (valid) => {
    if (!valid) return;

    loading.value = true;
    try {
      await authStore.register({
        phone: form.phone,
        smsCode: form.smsCode,
        nickname: form.nickname,
        password: form.password
      });

      ElMessage.success('注册成功');
      router.push('/dashboard');
    } catch (error: any) {
      ElMessage.error(error.message || '注册失败');
    } finally {
      loading.value = false;
    }
  });
};
</script>

<style scoped lang="scss">
.register-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  background: linear-gradient(160deg, #0a1628 0%, #1e293e 40%, #0d3b66 100%);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: -50%;
    right: -20%;
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba(0, 145, 255, 0.15) 0%, transparent 70%);
    border-radius: 50%;
  }
}

.register-box {
  width: 450px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.97);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 1;
}

.register-header {
  text-align: center;
  margin-bottom: 32px;

  h1 {
    font-size: 24px;
    font-weight: 600;
    color: #1d2129;
    margin-bottom: 8px;
  }

  p {
    font-size: 13px;
    color: #86909c;
    letter-spacing: 0.5px;
  }
}

.register-form {
  .el-form-item {
    margin-bottom: 24px;
  }
}

.sms-input-group {
  display: flex;
  gap: 8px;

  .el-input {
    flex: 1;
  }

  .el-button {
    width: 120px;
  }
}

.register-footer {
  text-align: center;
  margin-top: 24px;

  a {
    font-size: 14px;
    color: #0091FF;
    text-decoration: none;

    &:hover {
      color: #33a8ff;
    }
  }
}
</style>
