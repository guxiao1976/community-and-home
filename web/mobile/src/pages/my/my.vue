<template>
  <view class="page">
    <!-- Loading state -->
    <view v-if="pageLoading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- Unauthenticated state -->
    <view v-else-if="!userStore.isLoggedIn" class="login-prompt" @click="goLogin">
      <view class="login-avatar">
        <text class="login-avatar-emoji">👤</text>
      </view>
      <text class="login-text">点击登录</text>
    </view>

    <!-- Authenticated: V9 Card Layout -->
    <template v-else>
      <!-- Gradient Header -->
      <view class="header">
        <view class="avatar-circle">
          <text class="avatar-emoji">👤</text>
        </view>
        <text class="header-name">{{ userStore.nickname || '当前用户' }}</text>
        <text class="header-phone">{{ displayPhone }}</text>
      </view>

      <!-- Community Management Section -->
      <view class="section">
        <view class="section-header">
          <text class="section-title">🏘️ 小区管理</text>
          <text class="section-badge">已加入 {{ communityStore.communityCount }}/3</text>
        </view>
        <view class="card-row">
          <view class="action-card" hover-class="action-card--hover" @click="goJoinCommunity">
            <view class="action-icon-box action-icon-box--join">
              <text class="action-icon-emoji">➕</text>
            </view>
            <text class="action-label">加入小区</text>
          </view>
          <view class="action-card" hover-class="action-card--hover" @click="goLeaveCommunity">
            <view class="action-icon-box action-icon-box--leave">
              <text class="action-icon-emoji">🚪</text>
            </view>
            <text class="action-label">退出小区</text>
          </view>
        </view>
      </view>

      <!-- Settings Section -->
      <view class="section">
        <view class="section-header">
          <text class="section-title">⚙️ 设置</text>
        </view>
        <view class="settings-list">
          <view class="settings-item" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">个人信息</text>
            <text class="settings-item-arrow">→</text>
          </view>
          <view class="settings-item" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">账号安全</text>
            <text class="settings-item-arrow">→</text>
          </view>
          <view class="settings-item settings-item--last" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">关于我们</text>
            <text class="settings-item-arrow">→</text>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';

const userStore = useUserStore();
const communityStore = useCommunityStore();

const pageLoading = ref(true);

// Phone display with masking
const displayPhone = computed(() => {
  try {
    const phone = uni.getStorageSync('user_phone') as string;
    if (phone && phone.length >= 11) {
      return phone.slice(0, 3) + '****' + phone.slice(-4);
    }
  } catch {
    // ignore storage errors
  }
  return '未绑定手机号';
});

// Navigation
function goLogin() {
  uni.navigateTo({ url: '/pages/login/login' });
}

function goJoinCommunity() {
  uni.navigateTo({ url: '/pages/join-community/join-community' });
}

function goLeaveCommunity() {
  uni.navigateTo({ url: '/pages/leave-community/leave-community' });
}

function showDevToast() {
  uni.showToast({ title: '页面开发中', icon: 'none', duration: 1500 });
}

onMounted(async () => {
  // Ensure user profile is loaded (token may exist without user in store)
  if (isAuthenticated() && !userStore.user) {
    try {
      const user = await getUserProfile();
      userStore.setUser(user);
    } catch (e) {
      console.warn('[my] Failed to load user profile:', e);
    }
  }
  if (userStore.isLoggedIn) {
    await communityStore.loadMemberships();
  }
  pageLoading.value = false;
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
}

// ---- Loading ----
.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}

// ---- Login prompt ----
.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 200rpx 0;
}

.login-avatar {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 3rpx solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 4rpx 20rpx rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
}

.login-avatar-emoji {
  font-size: 72rpx;
}

.login-text {
  font-size: 32rpx;
  color: #B8956A;
}

// ---- Gradient Header ----
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 0 60rpx;
  background: linear-gradient(160deg, #D4B896 0%, #E8DCCF 30%, #FAF8F5 55%, #FFFFFF 80%);
}

.avatar-circle {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 3rpx solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 4rpx 20rpx rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20rpx;
}

.avatar-emoji {
  font-size: 72rpx;
}

.header-name {
  font-size: 24rpx;
  color: #A6988A;
  margin-bottom: 8rpx;
}

.header-phone {
  font-size: 32rpx;
  font-weight: 600;
  color: #3D3226;
}

// ---- Sections ----
.section {
  padding: 0 40rpx;
  margin-top: 32rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #3D3226;
}

.section-badge {
  font-size: 24rpx;
  color: #B8956A;
  background-color: rgba(184, 149, 106, 0.1);
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
}

// ---- Action Cards ----
.card-row {
  display: flex;
  gap: 20rpx;
}

.action-card {
  flex: 1;
  height: 260rpx;
  background-color: #FAF8F5;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(184, 149, 106, 0.08);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  &--hover {
    opacity: 0.85;
    transform: scale(0.97);
    transition: all 0.15s ease;
  }
}

.action-icon-box {
  width: 160rpx;
  height: 160rpx;
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16rpx;

  &--join {
    background: linear-gradient(135deg, #B8956A, #D4B896);
  }

  &--leave {
    background: linear-gradient(135deg, #D4958A, #E0ADA5);
  }
}

.action-icon-emoji {
  font-size: 84rpx;
}

.action-label {
  font-size: 28rpx;
  font-weight: 500;
  color: #3D3226;
}

// ---- Settings ----
.settings-list {
  background-color: #FAF8F5;
  border-radius: 20rpx;
  padding: 28rpx 32rpx;
}

.settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(184, 149, 106, 0.1);

  &--last {
    border-bottom: none;
  }

  &--hover {
    background-color: rgba(184, 149, 106, 0.04);
    transition: background-color 0.15s ease;
  }
}

.settings-item-label {
  font-size: 28rpx;
  color: #3D3226;
}

.settings-item-arrow {
  font-size: 24rpx;
  color: #CCC4BA;
}
</style>
