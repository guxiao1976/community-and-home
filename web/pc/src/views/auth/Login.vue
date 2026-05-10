<template>
  <div class="login-container">
    <div class="login-box">
      <div class="login-header">
        <h1>社区家园管理平台</h1>
        <p>Community & Home Management System</p>
      </div>

      <el-tabs v-model="activeTab" class="login-tabs">
        <el-tab-pane label="密码登录" name="password">
          <el-form
            ref="passwordFormRef"
            :model="passwordForm"
            :rules="passwordRules"
            class="login-form"
            @submit.prevent="handlePasswordLogin"
          >
            <el-form-item prop="phone">
              <el-input
                v-model="passwordForm.phone"
                placeholder="请输入手机号"
                prefix-icon="Phone"
                size="large"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input
                v-model="passwordForm.password"
                type="password"
                placeholder="请输入密码"
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
                @click="handlePasswordLogin"
              >
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="短信登录" name="sms">
          <el-form
            ref="smsFormRef"
            :model="smsForm"
            :rules="smsRules"
            class="login-form"
            @submit.prevent="handleSmsLogin"
          >
            <el-form-item prop="phone">
              <el-input
                v-model="smsForm.phone"
                placeholder="请输入手机号"
                prefix-icon="Phone"
                size="large"
              />
            </el-form-item>
            <el-form-item prop="smsCode">
              <div class="sms-input-group">
                <el-input
                  v-model="smsForm.smsCode"
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
            <el-form-item>
              <el-button
                type="primary"
                size="large"
                :loading="loading"
                style="width: 100%"
                @click="handleSmsLogin"
              >
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>

      <div class="login-footer">
        <router-link to="/register">还没有账号？立即注册</router-link>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ElMessage, type FormInstance } from 'element-plus';
import { useAuthStore } from '@/stores/auth';
import { sendSms } from '@/api/identity';
import { requiredPhoneRule, requiredPasswordRule, smsCodeRule } from '@/utils/validation';

const router = useRouter();
const route = useRoute();
const authStore = useAuthStore();

const activeTab = ref('password');
const loading = ref(false);
const smsCountdown = ref(0);

// Password login form
const passwordFormRef = ref<FormInstance>();
const passwordForm = reactive({
  phone: '',
  password: ''
});

const passwordRules = {
  phone: requiredPhoneRule,
  password: requiredPasswordRule
};

// SMS login form
const smsFormRef = ref<FormInstance>();
const smsForm = reactive({
  phone: '',
  smsCode: ''
});

const smsRules = {
  phone: requiredPhoneRule,
  smsCode: smsCodeRule
};

// Handle password login
const handlePasswordLogin = async () => {
  if (!passwordFormRef.value) return;

  await passwordFormRef.value.validate(async (valid) => {
    if (!valid) return;

    loading.value = true;
    try {
      await authStore.login(passwordForm.phone, passwordForm.password);
      ElMessage.success('登录成功');

      // Redirect to original page or dashboard
      const redirect = (route.query.redirect as string) || '/dashboard';
      router.push(redirect);
    } catch (error: any) {
      ElMessage.error(error.message || '登录失败');
    } finally {
      loading.value = false;
    }
  });
};

// Handle SMS login
const handleSmsLogin = async () => {
  if (!smsFormRef.value) return;

  await smsFormRef.value.validate(async (valid) => {
    if (!valid) return;

    loading.value = true;
    try {
      await authStore.loginWithSms(smsForm.phone, smsForm.smsCode);
      ElMessage.success('登录成功');

      const redirect = (route.query.redirect as string) || '/dashboard';
      router.push(redirect);
    } catch (error: any) {
      ElMessage.error(error.message || '登录失败');
    } finally {
      loading.value = false;
    }
  });
};

// Send SMS code
const handleSendSms = async () => {
  if (!smsForm.phone) {
    ElMessage.warning('请输入手机号');
    return;
  }

  if (!/^1[3-9]\d{9}$/.test(smsForm.phone)) {
    ElMessage.warning('请输入正确的手机号');
    return;
  }

  try {
    await sendSms(smsForm.phone);
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
</script>

<style scoped lang="scss">
.login-container {
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

  &::after {
    content: '';
    position: absolute;
    bottom: -30%;
    left: -10%;
    width: 400px;
    height: 400px;
    background: radial-gradient(circle, rgba(0, 145, 255, 0.1) 0%, transparent 70%);
    border-radius: 50%;
  }
}

.login-box {
  width: 400px;
  padding: 40px;
  background: rgba(255, 255, 255, 0.97);
  border-radius: 12px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 1;
}

.login-header {
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

.login-tabs {
  :deep(.el-tabs__nav-wrap::after) {
    display: none;
  }

  :deep(.el-tabs__item) {
    font-size: 15px;
    color: #86909c;

    &.is-active {
      color: #0091FF;
    }
  }

  :deep(.el-tabs__active-bar) {
    background-color: #0091FF;
  }
}

.login-form {
  margin-top: 24px;

  .el-form-item {
    margin-bottom: 24px;
  }

  :deep(.el-input__wrapper) {
    border-radius: 8px;
    padding: 4px 12px;
    box-shadow: 0 0 0 1px #e5e6eb inset;

    &:hover {
      box-shadow: 0 0 0 1px #0091FF inset;
    }

    &.is-focus {
      box-shadow: 0 0 0 1px #0091FF inset;
    }
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
    border-radius: 8px;
  }
}

.login-footer {
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
