<template>
  <view class="page">
    <!-- 加载中 -->
    <view v-if="pageLoading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 未登录 -->
    <view v-else-if="!userStore.isLoggedIn" class="login-prompt" @click="goLogin">
      <view class="avatar-placeholder">
        <text class="avatar-icon">👤</text>
      </view>
      <text class="login-text">点击登录</text>
    </view>

    <!-- 已登录：左侧菜单 + 右侧内容 -->
    <template v-else>
      <view class="layout">
        <!-- 左侧菜单栏 -->
        <scroll-view class="menu-bar" scroll-y>
          <view
            v-for="menu in menus"
            :key="menu.key"
            class="menu-item"
            :class="{ 'menu-item--active': activeMenu === menu.key }"
            @click="activeMenu = menu.key"
          >
            <text class="menu-item-icon">{{ menu.icon }}</text>
            <text class="menu-item-label">{{ menu.label }}</text>
          </view>
        </scroll-view>

        <!-- 右侧内容区 -->
        <scroll-view class="content-area" scroll-y>
          <!-- 用户信息卡片 -->
          <view class="user-card">
            <view class="avatar">
              <image
                v-if="userStore.avatar"
                :src="userStore.avatar"
                class="avatar-img"
                mode="aspectFill"
              />
              <text v-else class="avatar-icon">👤</text>
            </view>
            <view class="user-info">
              <text class="nickname">{{ userStore.nickname || '用户' }}</text>
              <text class="community-count">
                已加入 {{ communityStore.communityCount }}/3 个小区
              </text>
            </view>
          </view>

          <!-- 退出小区别面板 -->
          <view v-if="activeMenu === 'leave'" class="panel">
            <view class="panel-header">
              <text class="panel-title">退出小区</text>
              <text class="panel-subtitle">选择要退出的小区</text>
            </view>

            <view v-if="communityStore.communities.length === 0" class="panel-empty">
              <text class="panel-empty-icon">🏘️</text>
              <text class="panel-empty-text">暂未加入任何小区</text>
            </view>

            <view v-else class="community-list">
              <view
                v-for="c in communityStore.communities"
                :key="c.communityId"
                class="community-card"
                :class="{ 'community-card--leaving': leavingId === c.communityId }"
                @click="confirmLeave(c)"
              >
                <view class="community-card-left">
                  <text class="community-card-icon">🏘️</text>
                  <view class="community-card-info">
                    <text class="community-card-name">{{ c.communityName }}</text>
                    <text v-if="c.address" class="community-card-addr">{{ c.address }}</text>
                  </view>
                </view>
                <view class="community-card-right">
                  <text v-if="leavingId === c.communityId" class="community-card-hint">确认退出?</text>
                  <text v-else class="community-card-action">退出</text>
                </view>
              </view>
            </view>
          </view>

          <!-- 其他菜单占位 -->
          <view v-else class="panel">
            <view class="panel-empty">
              <text class="panel-empty-icon">🔧</text>
              <text class="panel-empty-text">功能开发中...</text>
            </view>
          </view>
        </scroll-view>
      </view>

      <!-- 确认弹窗 -->
      <view v-if="confirmTarget" class="modal-mask" @click="cancelLeave">
        <view class="modal-card" @click.stop>
          <text class="modal-icon">⚠️</text>
          <text class="modal-title">确认退出？</text>
          <text class="modal-desc">
            退出「{{ confirmTarget.communityName }}」后，将不再接收该小区的通知公告等信息。
          </text>
          <view class="modal-actions">
            <view class="modal-btn modal-btn--cancel" @click="cancelLeave">
              <text>取消</text>
            </view>
            <view class="modal-btn modal-btn--confirm" @click="doLeave">
              <text>确认退出</text>
            </view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import type { CommunityInfo } from '@/stores/community';
import { leaveCommunity } from '@/api/user';
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';

const userStore = useUserStore();
const communityStore = useCommunityStore();

// ---- Menu ----
const menus = [
  { key: 'leave', icon: '🚪', label: '退出小区' },
  { key: 'settings', icon: '⚙️', label: '设置' },
  { key: 'about', icon: 'ℹ️', label: '关于' },
];

const activeMenu = ref('leave');

// ---- Leave Community State ----
const leavingId = ref('');
const confirmTarget = ref<CommunityInfo | null>(null);

function confirmLeave(c: CommunityInfo) {
  if (leavingId.value === c.communityId) {
    confirmTarget.value = c;
    leavingId.value = '';
  } else {
    leavingId.value = c.communityId;
  }
}

function cancelLeave() {
  confirmTarget.value = null;
  leavingId.value = '';
}

async function doLeave() {
  if (!confirmTarget.value) return;
  const target = confirmTarget.value;
  try {
    uni.showLoading({ title: '退出中...', mask: true });
    await leaveCommunity(target.communityId);
    communityStore.removeCommunity(target.communityId);
    uni.hideLoading();
    uni.showToast({ title: '已退出小区', icon: 'success', duration: 1500 });
  } catch {
    uni.hideLoading();
    uni.showToast({ title: '退出失败，请重试', icon: 'none', duration: 2000 });
  }
  confirmTarget.value = null;
  leavingId.value = '';
}

function goLogin() {
  uni.navigateTo({ url: '/pages/login/login' });
}

const pageLoading = ref(true);

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
  background-color: $uni-bg-color;
}

.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}

.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 200rpx 0;
}

.avatar-placeholder {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  background-color: $uni-bg-color-grey;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20rpx;
}

.avatar-icon {
  font-size: 60rpx;
}

.login-text {
  font-size: 32rpx;
  color: $uni-color-primary;
}

.layout {
  display: flex;
  height: 100vh;
}

.menu-bar {
  width: 180rpx;
  background-color: $uni-bg-color-card;
  border-right: 1px solid $uni-border-color;
  flex-shrink: 0;
  padding-top: 16rpx;
}

.menu-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 28rpx 8rpx;
  position: relative;

  &--active {
    background: linear-gradient(90deg, rgba($uni-color-primary, 0.08), transparent);

    &::before {
      content: '';
      position: absolute;
      left: 0;
      top: 16rpx;
      bottom: 16rpx;
      width: 6rpx;
      background: $uni-color-primary;
      border-radius: 0 3rpx 3rpx 0;
    }

    .menu-item-label {
      color: $uni-color-primary;
      font-weight: 600;
    }
  }
}

.menu-item-icon {
  font-size: 40rpx;
  margin-bottom: 8rpx;
}

.menu-item-label {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

.content-area {
  flex: 1;
  background-color: $uni-bg-color;
}

.user-card {
  display: flex;
  align-items: center;
  padding: 32rpx;
  background: linear-gradient(135deg, rgba($uni-color-primary-light, 0.2), $uni-bg-color-card);
  margin: 20rpx;
  border-radius: 16rpx;
}

.avatar {
  width: 88rpx;
  height: 88rpx;
  border-radius: 50%;
  background-color: $uni-bg-color-grey;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  margin-right: 20rpx;
  flex-shrink: 0;
}

.avatar-img {
  width: 100%;
  height: 100%;
}

.avatar-icon {
  font-size: 44rpx;
}

.user-info {
  flex: 1;
}

.nickname {
  display: block;
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 4rpx;
}

.community-count {
  font-size: 24rpx;
  color: $uni-color-primary;
}

.panel {
  padding: 0 20rpx 40rpx;
}

.panel-header {
  padding: 20rpx 12rpx 20rpx;
}

.panel-title {
  display: block;
  font-size: 34rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 4rpx;
}

.panel-subtitle {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

.panel-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 0;
}

.panel-empty-icon {
  font-size: 80rpx;
  margin-bottom: 16rpx;
  opacity: 0.5;
}

.panel-empty-text {
  font-size: 26rpx;
  color: $uni-text-color-placeholder;
}

.community-list {
  display: flex;
  flex-direction: column;
  gap: 12rpx;
}

.community-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: $uni-bg-color-card;
  border-radius: 12rpx;
  padding: 20rpx;
  box-shadow: $uni-shadow-sm;

  &--leaving {
    border: 1px solid $uni-color-error;
  }
}

.community-card-left {
  display: flex;
  align-items: center;
  flex: 1;
  min-width: 0;
}

.community-card-icon {
  font-size: 36rpx;
  margin-right: 14rpx;
  flex-shrink: 0;
}

.community-card-info {
  flex: 1;
  min-width: 0;
}

.community-card-name {
  display: block;
  font-size: 28rpx;
  font-weight: 500;
  color: $uni-text-color;
}

.community-card-addr {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 2rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.community-card-right {
  flex-shrink: 0;
  margin-left: 16rpx;
}

.community-card-action {
  font-size: 24rpx;
  color: $uni-color-error;
  padding: 8rpx 20rpx;
  border-radius: 20rpx;
  background: rgba($uni-color-error, 0.08);
}

.community-card-hint {
  font-size: 24rpx;
  color: $uni-color-error;
  font-weight: 600;
}

.modal-mask {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-card {
  width: 560rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 48rpx 40rpx 36rpx;
  text-align: center;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.1);
}

.modal-icon {
  font-size: 72rpx;
  display: block;
  margin-bottom: 16rpx;
}

.modal-title {
  display: block;
  font-size: 34rpx;
  font-weight: 700;
  color: $uni-text-color;
  margin-bottom: 12rpx;
}

.modal-desc {
  display: block;
  font-size: 26rpx;
  color: $uni-text-color-grey;
  line-height: 1.6;
  margin-bottom: 32rpx;
}

.modal-actions {
  display: flex;
  gap: 20rpx;
}

.modal-btn {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: 600;

  &--cancel {
    background: $uni-bg-color-grey;
    color: $uni-text-color-grey;
  }

  &--confirm {
    background: linear-gradient(135deg, $uni-color-error, #C07A70);
    color: #fff;
  }
}
</style>
