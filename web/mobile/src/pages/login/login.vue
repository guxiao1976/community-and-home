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

      <!-- Agreement Checkbox -->
      <view class="agreement" @click="toggleAgreement">
        <view class="checkbox" :class="{ 'checkbox--checked': agreed }">
          <text v-if="agreed" class="check-mark">✓</text>
        </view>
        <view class="agreement-text">
          已阅并同意<text class="agreement-link" @click.stop="showAgreement">《使用协议》</text>
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
import { loginWithSms, register, sendSmsCode, getUserProfile, ensurePublicKey } from '@/api/identity';
import { getUserMemberships } from '@/api/user';
import { useUserStore } from '@/stores/user';
import { getDeviceId } from '@/utils/device';

const userStore = useUserStore();

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
const agreed = ref(false);
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
const canSubmit = computed(
  () => isPhoneValid.value && isCodeValid.value && agreed.value,
);

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

// --- Agreement ---
function toggleAgreement() {
  agreed.value = !agreed.value;
}

function showAgreement() {
  uni.showToast({ title: '使用协议页面开发中', icon: 'none', duration: 2000 });
}

// --- Submit ---
async function handleSubmit() {
  if (!canSubmit.value || submitting.value) return;

  submitting.value = true;
  const deviceId = getDeviceId();
  const phoneValue = phone.value;
  const codeValue = smsCode.value;

  // Step 1: Try SMS login first
  try {
    const loginRes = await loginWithSms(phoneValue, codeValue, deviceId);
    await onAuthSuccess(loginRes);
    submitting.value = false;
    return;
  } catch (loginErr: unknown) {
    const msg = (loginErr as Error)?.message || '';
    // Only auto-register if user doesn't exist (error 50001)
    if (!msg.includes('50001') && !msg.includes('未注册')) {
      submitting.value = false;
      return; // Wrong code or other error — stop, toast already shown by interceptor
    }
  }

  // Step 2: Auto-register (user doesn't exist)
  const autoNickname = '用户' + phoneValue.slice(-4);
  try {
    uni.showLoading({ title: '注册中...', mask: true });
    const regRes = await register({
      phone: phoneValue,
      smsCode: codeValue,
      nickname: autoNickname,
      deviceId,
    });
    uni.hideLoading();
    await onAuthSuccess(regRes);
    submitting.value = false;
  } catch (err: unknown) {
    uni.hideLoading();
    submitting.value = false;
    console.error('[Login] Register failed:', err);
  }
}

async function onAuthSuccess(loginRes: {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}) {
  // 1. Save tokens
  userStore.setAuth(loginRes);

  // 2. Fetch user profile
  try {
    uni.showLoading({ title: '登录中...', mask: true });
    const user = await getUserProfile();
    userStore.setUser(user);
  } catch (_err) {
    console.warn('[Login] Failed to fetch user profile, continuing anyway');
  } finally {
    uni.hideLoading();
  }

  // 3. Check if user has communities, then navigate accordingly
  let hasCommunities = false;
  try {
    const memberships = await getUserMemberships();
    hasCommunities = memberships.length > 0;
  } catch {
    // If membership check fails, default to join-community
  }

  uni.showToast({ title: '登录成功', icon: 'success', duration: 1500 });
  setTimeout(() => {
    if (hasCommunities) {
      uni.switchTab({ url: '/pages/notice/notice' });
    } else {
      uni.redirectTo({ url: '/pages/join-community/join-community' });
    }
    submitting.value = false;
  }, 800);
}
</script>

<style scoped lang="scss">
.login-page {
  min-height: 100vh;
  background: linear-gradient(160deg, #B8956A 0%, #D4B896 30%, #FFFFFF 30%);
  padding: 0 32px;
}

/* --- Header --- */
.header {
  padding-top: 120rpx;
  padding-bottom: 60rpx;
  text-align: center;

  .logo-area {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 12rpx;
    margin-bottom: 16rpx;

    .logo-icon {
      font-size: 72rpx;
    }

    .app-name {
      font-size: 52rpx;
      font-weight: 700;
      color: #fff;
    }
  }

  .welcome-text {
    display: block;
    font-size: 28rpx;
    color: rgba(255, 255, 255, 0.85);
  }
}

/* --- Form --- */
.form {
  background: #fff;
  border-radius: 20rpx;
  padding: 48rpx 36rpx;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.06);
}

/* --- Input Group --- */
.input-group {
  display: flex;
  align-items: center;
  background: $uni-bg-color-grey;
  border-radius: 12rpx;
  padding: 0 24rpx;
  margin-bottom: 24rpx;
  height: 96rpx;
  position: relative;

  .input-icon {
    font-size: 34rpx;
    margin-right: 16rpx;
    flex-shrink: 0;
  }

  .input-field {
    flex: 1;
    font-size: 28rpx;
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
    width: 44rpx;
    height: 44rpx;
    border-radius: 50%;
    background: #c8c9cc;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    text {
      color: #fff;
      font-size: 22rpx;
      font-weight: 700;
    }
  }

  .sms-btn {
    flex-shrink: 0;
    padding: 12rpx 24rpx;
    border-radius: 8rpx;
    background: $uni-color-primary;
    font-size: 24rpx;
    color: #fff;
    white-space: nowrap;

    &--disabled {
      background: #c8c9cc;
      color: #fff;
    }
  }
}

/* --- Agreement --- */
.agreement {
  display: flex;
  align-items: center;
  margin-bottom: 32rpx;
  padding: 8rpx 0;

  .checkbox {
    width: 36rpx;
    height: 36rpx;
    border-radius: 6rpx;
    border: 2rpx solid #c8c9cc;
    margin-right: 12rpx;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
    transition: all 0.2s;

    &--checked {
      background: $uni-color-primary;
      border-color: $uni-color-primary;

      .check-mark {
        color: #fff;
        font-size: 24rpx;
        font-weight: 700;
      }
    }
  }

  .agreement-text {
    display: flex;
    align-items: center;
    font-size: 24rpx;
    color: $uni-text-color-grey;

    .agreement-link {
      color: $uni-color-primary;
      text-decoration: underline;
    }
  }
}

/* --- Submit Button --- */
.submit-btn {
  width: 100%;
  height: 96rpx;
  border-radius: 48rpx;
  background: linear-gradient(135deg, #B8956A, #D4B896);
  color: #fff;
  font-size: 32rpx;
  font-weight: 600;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  margin-bottom: 24rpx;

  &::after {
    border: none;
  }

  &--disabled {
    background: #c8c9cc;
    color: #fff;
  }
}
</style>
