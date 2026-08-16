<template>
  <view class="page">
    <!-- Loading -->
    <view v-if="pageLoading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- Not logged in -->
    <view v-else-if="!userStore.isLoggedIn" class="login-prompt" @click="goLogin">
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
import { ref, onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';

const userStore = useUserStore();
const communityStore = useCommunityStore();

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
      console.warn('[mine] Failed to load user profile:', e);
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
  padding: 2rem 1.5rem;
}

.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 6.25rem 0;
}

.loading-text {
  font-size: 0.875rem;
  color: $uni-text-color-placeholder;
}

.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3.75rem 0;

  .avatar-placeholder {
    width: 3.75rem;
    height: 3.75rem;
    border-radius: 50%;
    background-color: $uni-bg-color-grey;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.625rem;

    .avatar-icon {
      font-size: 1.875rem;
    }
  }

  .login-text {
    font-size: 1rem;
    color: $uni-color-primary;
  }
}

.profile {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1.875rem 0;

  .avatar {
    width: 3.75rem;
    height: 3.75rem;
    border-radius: 50%;
    background-color: $uni-bg-color-grey;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 0.625rem;
    overflow: hidden;

    .avatar-img {
      width: 100%;
      height: 100%;
    }

    .avatar-icon {
      font-size: 1.875rem;
    }
  }

  .nickname {
    font-size: 1.125rem;
    font-weight: 600;
    color: $uni-text-color;
  }
}

.community-info {
  margin-top: 0.375rem;
}

.community-count {
  font-size: 0.8125rem;
  color: $uni-color-primary;
}

.menu-section {
  margin-top: 1.5rem;
  display: flex;
  justify-content: center;
}

.menu-placeholder {
  font-size: 0.875rem;
  color: $uni-text-color-placeholder;
}
</style>
