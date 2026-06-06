<template>
  <view class="page">
    <!-- Not logged in -->
    <view v-if="!userStore.isLoggedIn" class="login-prompt" @click="goLogin">
      <view class="avatar-placeholder">
        <text class="avatar-icon">👤</text>
      </view>
      <text class="login-text">点击登录</text>
    </view>

    <!-- Logged in -->
    <view v-else class="profile">
      <view class="avatar">
        <image
          v-if="userStore.avatar"
          :src="userStore.avatar"
          class="avatar-img"
          mode="aspectFill"
        />
        <text v-else class="avatar-icon">👤</text>
      </view>
      <text class="nickname">{{ userStore.nickname }}</text>
      <view class="community-info" v-if="communityStore.communityCount > 0">
        <text class="community-count">已加入 {{ communityStore.communityCount }}/3 个小区</text>
      </view>
      <view class="community-info" v-else>
        <text class="community-count">暂未加入小区</text>
      </view>
    </view>

    <!-- Menu placeholder -->
    <view class="menu-section">
      <text class="menu-placeholder">更多功能开发中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';

const userStore = useUserStore();
const communityStore = useCommunityStore();

function goLogin() {
  uni.navigateTo({ url: '/pages/login/login' });
}

onMounted(async () => {
  // Ensure user profile is loaded (token may exist without user in store)
  if (isAuthenticated() && !userStore.user) {
    try {
      const user = await getUserProfile();
      userStore.setUser(user);
    } catch { /* ignore */ }
  }
  if (userStore.isLoggedIn) {
    communityStore.loadMemberships();
  }
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  padding: 32px 24px;
}

.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 0;

  .avatar-placeholder {
    width: 120rpx;
    height: 120rpx;
    border-radius: 50%;
    background-color: $uni-bg-color-grey;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 20rpx;

    .avatar-icon {
      font-size: 60rpx;
    }
  }

  .login-text {
    font-size: 32rpx;
    color: $uni-color-primary;
  }
}

.profile {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 60rpx 0;

  .avatar {
    width: 120rpx;
    height: 120rpx;
    border-radius: 50%;
    background-color: $uni-bg-color-grey;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 20rpx;
    overflow: hidden;

    .avatar-img {
      width: 100%;
      height: 100%;
    }

    .avatar-icon {
      font-size: 60rpx;
    }
  }

  .nickname {
    font-size: 36rpx;
    font-weight: 600;
    color: $uni-text-color;
  }
}

.community-info {
  margin-top: 12rpx;
}

.community-count {
  font-size: 26rpx;
  color: $uni-color-primary;
}

.menu-section {
  margin-top: 48rpx;
  display: flex;
  justify-content: center;
}

.menu-placeholder {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}
</style>
