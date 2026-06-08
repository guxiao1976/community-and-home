<template>
  <view class="page">
    <!-- 导航栏 -->
    <view class="nav-bar">
      <view class="nav-back" @click="goBack">← 返回</view>
      <text class="nav-title">账号安全</text>
    </view>

    <!-- 手机号 -->
    <view class="section">
      <view class="card">
        <view class="card-row">
          <text class="row-label">当前手机号</text>
          <text class="row-value">{{ phone }}</text>
        </view>
      </view>
    </view>

    <!-- 退出登录 -->
    <view class="section">
      <view class="btn-logout" @click="handleLogout">退出登录</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useUserStore } from '@/stores/user';

const userStore = useUserStore();

const phone = computed(() => {
  return userStore.user?.phone || uni.getStorageSync('user_phone') || '未绑定';
});

function goBack() {
  uni.navigateBack();
}

function handleLogout() {
  uni.showModal({
    title: '确认退出',
    content: '退出后需要重新登录',
    success: (res) => {
      if (res.confirm) {
        userStore.logout();
        uni.reLaunch({ url: '/pages/login/login' });
      }
    },
  });
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #FFFFFF;
  padding: 0 32rpx;
}

.nav-bar {
  display: flex;
  align-items: center;
  padding: 28rpx 0;
  border-bottom: 1rpx solid #F0EBE3;
}

.nav-back {
  font-size: 28rpx;
  color: #B8956A;
  margin-right: 16rpx;
}

.nav-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #3D3226;
}

.section {
  margin-top: 32rpx;
}

.card {
  background: #FAF8F5;
  border-radius: 16rpx;
  padding: 28rpx 32rpx;
}

.card-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 14rpx 0;
}

.row-label {
  font-size: 28rpx;
  color: #3D3226;
}

.row-value {
  font-size: 28rpx;
  color: #A6988A;
}

.btn-logout {
  height: 88rpx;
  border-radius: 44rpx;
  background: linear-gradient(135deg, #D4958A, #C1786E);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32rpx;
  font-weight: 600;
  color: #fff;
}
</style>
