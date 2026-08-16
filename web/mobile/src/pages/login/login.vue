<template>
  <view class="login-page">
    <!-- Header -->
    <view class="header">
      <view class="logo-area">
        <text class="logo-icon">🏠</text>
        <text class="app-name">社区家园</text>
      </view>
      <text class="welcome-text">欢迎来到社区家园</text>
    </view>

    <!-- Form -->
    <view class="form">
      <!-- Phone Input -->
      <view class="input-group">
        <view class="input-icon">📱</view>
        <input
          v-model="phone"
          class="input-field"
          type="number"
          placeholder="请输入手机号"
          maxlength="11"
          :disabled="submitting"
          @input="onPhoneInput"
        />
        <view v-if="phone" class="clear-btn" @click="clearPhone">
          <text>✕</text>
        </view>
      </view>

      <!-- SMS Code Input -->
      <view class="input-group">
        <view class="input-icon">✉️</view>
        <input
          v-model="smsCode"
          class="input-field"
          type="number"
          placeholder="请输入验证码"
          maxlength="6"
          :disabled="submitting"
        />
        <view
          class="sms-btn"
          :class="{ 'sms-btn--disabled': !canSend || countdown > 0 }"
          @click="sendCode"
        >
          <text v-if="countdown === 0">获取验证码</text>
          <text v-else>{{ countdown }}s</text>
        </view>
      </view>

      <!-- Submit Button -->
      <button
        class="submit-btn"
        :class="{ 'submit-btn--disabled': !canSubmit }"
        :disabled="!canSubmit || submitting"
        :loading="submitting"
        @click="handleSubmit"
      >
        {{ submitting ? '请稍候...' : '登录/注册' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { loginWithSms, sendSmsCode, ensurePublicKey } from '@/api/identity';
import { getDeviceId } from '@/utils/device';
import { handleAuthSuccess } from '@/utils/auth-flow';
// SEE: [[sms-code-persist-localstorage]] — 一次性验证码经共享模块走内存态 + sessionStorage，不落 localStorage
// SEE: [[frontend-cross-page-storage-contract]] — key/结构收敛到共享契约源，不再内联 magic string
// SEE: [[cross-page-sensitive-temp-data-storage]] — 跨页一次性敏感数据优先内存态载体
import { saveRegPending } from '@/utils/reg-pending';

// 预加载 RSA 公钥，加快首次登录响应速度
onMounted(async () => {
  try {
    await ensurePublicKey();
  } catch (_err) {
    // 公钥加载失败不阻塞页面，登录时重试
    console.warn('[Login] Preload public key failed, will retry on submit');
  }
});

// --- Form State ---
const phone = ref('');
const smsCode = ref('');
const submitting = ref(false);

// --- SMS Countdown ---
const countdown = ref(0);
let countdownTimer: ReturnType<typeof setInterval> | null = null;

onUnmounted(() => {
  if (countdownTimer) {
    clearInterval(countdownTimer);
  }
});

// --- Validation ---
const isPhoneValid = computed(() => /^1[3-9]\d{9}$/.test(phone.value));
const isCodeValid = computed(() => /^\d{4,6}$/.test(smsCode.value));
const canSend = computed(() => isPhoneValid.value && countdown.value === 0);
// 登录页 canSubmit 仅依赖手机号+验证码合法；协议确认已移至注册协议页（agreement.vue）
const canSubmit = computed(() => isPhoneValid.value && isCodeValid.value);

// --- Phone Input ---
function onPhoneInput() {
  phone.value = phone.value.replace(/\D/g, '');
}
function clearPhone() {
  phone.value = '';
}

// --- Send SMS ---
async function sendCode() {
  if (!canSend.value || countdown.value > 0) return;

  // 点击后立即启动 30s 最小冷却，防止连续点击
  countdown.value = 30;
  countdownTimer = setInterval(() => {
    countdown.value--;
    if (countdown.value <= 0) {
      if (countdownTimer) {
        clearInterval(countdownTimer);
        countdownTimer = null;
      }
    }
  }, 1000);

  try {
    uni.showLoading({ title: '发送中...' });
    await sendSmsCode(phone.value);
    uni.hideLoading();
    uni.showToast({ title: '验证码已发送', icon: 'success', duration: 1500 });
    // 成功后延长冷却至 60s，与后端限流一致
    countdown.value = 60;
  } catch (err: unknown) {
    uni.hideLoading();
    console.error('[Login] Send SMS failed:', err);
    // 保持 30s 冷却（countdown 已在点击时启动，不清零）
  }
}

// --- Submit ---
async function handleSubmit() {
  if (!canSubmit.value || submitting.value) return;

  submitting.value = true;
  const deviceId = getDeviceId();
  const phoneValue = phone.value;
  const codeValue = smsCode.value;

  // Step 1: 先试短信登录
  try {
    const loginRes = await loginWithSms(phoneValue, codeValue, deviceId);
    // 登录成功 → 共享登录后流程（保存 token + profile + 小区跳转）。
    // phone 传入 handleAuthSuccess 写入 user_phone（my.vue displayPhone 兜底）。
    // submitting 由 handleAuthSuccess 的 onCompleted 在跳转完成后复位，防止跳转窗口期二次点击触发重复登录
    await handleAuthSuccess(loginRes, {
      phone: phoneValue,
      onCompleted: () => {
        submitting.value = false;
      },
    });
    return;
  } catch (loginErr: unknown) {
    const e = loginErr as Error & { code?: number };
    const msg = e?.message || '';
    // 仅当 50001（用户未注册）进入协议注册流程；其他错误保持现有处理（拦截器已 toast）。
    // err.code 数值为主判据，msg 字符串匹配仅作旧后端兜底（code 缺失时）
    const isNotRegistered = e?.code !== undefined
      ? e.code === 50001
      : msg.includes('50001') || msg.includes('未注册');
    if (!isNotRegistered) {
      submitting.value = false;
      return;
    }
  }

  // Step 2: 50001 未注册 → 暂存注册所需数据，跳协议页确认注册（协议页读 storage → register → 自动登录）
  const regPending = {
    phone: phoneValue,
    smsCode: codeValue,
    deviceId,
    nickname: '用户' + phoneValue.slice(-4),
  };
  try {
    saveRegPending(regPending);
  } catch (err: unknown) {
    console.error('[Login] 暂存注册数据失败', err);
    submitting.value = false;
    uni.showToast({ title: '注册信息保存失败，请重试', icon: 'none' });
    return;
  }
  submitting.value = false;
  uni.navigateTo({ url: '/pages/agreement/agreement' });
}
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  background: linear-gradient(160deg, #B8956A 0%, #D4B896 30%, #FFFFFF 30%);
  padding: 0 2rem;
}

/* --- Header --- */
.header {
  padding-top: 3.75rem;
  padding-bottom: 1.875rem;
  text-align: center;

  .logo-area {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.375rem;
    margin-bottom: 0.5rem;

    .logo-icon {
      font-size: 2.25rem;
    }

    .app-name {
      font-size: 1.625rem;
      font-weight: 700;
      color: #fff;
    }
  }

  .welcome-text {
    display: block;
    font-size: 0.875rem;
    color: rgba(255, 255, 255, 0.85);
  }
}

/* --- Form --- */
.form {
  background: #fff;
  border-radius: 0.625rem;
  padding: 1.5rem 1.125rem;
  box-shadow: 0 0.25rem 1.25rem rgba(0, 0, 0, 0.06);
}

/* --- Input Group --- */
.input-group {
  display: flex;
  align-items: center;
  background: $uni-bg-color-grey;
  border-radius: 0.375rem;
  padding: 0 0.75rem;
  margin-bottom: 0.75rem;
  height: 3rem;
  position: relative;

  .input-icon {
    font-size: 1.0625rem;
    margin-right: 0.5rem;
    flex-shrink: 0;
  }

  .input-field {
    flex: 1;
    font-size: 0.875rem;
    color: $uni-text-color;
    height: 100%;
    background: transparent;

    &::placeholder {
      color: $uni-text-color-placeholder;
    }

    &:disabled {
      opacity: 0.6;
    }
  }

  .clear-btn {
    width: 1.375rem;
    height: 1.375rem;
    border-radius: 50%;
    background: #c8c9cc;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    text {
      color: #fff;
      font-size: 0.6875rem;
      font-weight: 700;
    }
  }

  .sms-btn {
    flex-shrink: 0;
    padding: 0.375rem 0.75rem;
    border-radius: 0.25rem;
    background: $uni-color-primary;
    font-size: 0.75rem;
    color: #fff;
    white-space: nowrap;

    &--disabled {
      background: #c8c9cc;
      color: #fff;
    }
  }
}

/* --- Submit Button --- */
.submit-btn {
  width: 100%;
  height: 3rem;
  border-radius: 1.5rem;
  background: linear-gradient(135deg, #B8956A, #D4B896);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  margin-bottom: 0.75rem;

  &::after {
    border: none;
  }

  &--disabled {
    background: #c8c9cc;
    color: #fff;
  }
}
</style>
