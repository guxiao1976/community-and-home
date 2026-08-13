<template>
  <view class="page">
    <view v-if="loading" class="state">
      <text class="state-text">加载中...</text>
    </view>

    <view v-else-if="error" class="state">
      <text class="state-text state-text--error">{{ error }}</text>
    </view>

    <view v-else-if="user" class="profile">
      <view class="avatar">
        <text class="avatar-emoji">👤</text>
      </view>
      <text class="nickname">{{ user.nickname || '用户' }}</text>
      <text class="phone">{{ user.phone || '未绑定手机号' }}</text>

      <!-- 同屋互见：后端 same_house=true 才返回楼/单元/房号与真实手机号，前端仅展示 -->
      <view v-if="sameHouseInfo?.same_house" class="house-card">
        <text class="house-label">同屋成员</text>
        <text class="house-addr">{{ sameHouseInfo.building }} 栋 {{ sameHouseInfo.unit }} 单元 {{ sameHouseInfo.room }} 室</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getUser, type SameHouseInfo, type GetUserDetailResult } from '@/api/user';
import { useUserStore } from '@/stores/user';

// 用户详情页：消费 GetUserResponse.same_house。
// phone 已由后端按 viewer 上下文决定明文/脱敏，前端不做二次脱敏（权威在后端）。
// SEE: [[web-common-type-reuse-no-redefine]]（user 类型复用 api/user.ts 的 GetUserDetailResult，不重复定义）
const userStore = useUserStore();

const loading = ref(true);
const error = ref('');
const user = ref<GetUserDetailResult['user'] | null>(null);
const sameHouseInfo = ref<SameHouseInfo | null>(null);

onLoad((options) => {
  load(options);
});

async function load(options: Record<string, string | undefined> | undefined) {
  const id = options?.id;
  if (!id) {
    error.value = '缺少用户 ID';
    loading.value = false;
    return;
  }
  // viewer_id 缺省 → 后端对手机号脱敏、无房屋号
  const viewerId = options?.viewer_id || userStore.userId || undefined;
  try {
    const res = await getUser(id, viewerId);
    user.value = res.user;
    sameHouseInfo.value = res.same_house || null;
  } catch (e: any) {
    error.value = e?.message || '加载失败';
  } finally {
    loading.value = false;
  }
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
}

.state {
  display: flex;
  justify-content: center;
  padding: 200rpx 0;

  .state-text {
    font-size: 28rpx;
    color: $uni-text-color-placeholder;

    &--error {
      color: $uni-color-error;
    }
  }
}

.profile {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 100rpx 40rpx;
}

.avatar {
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

.avatar-emoji {
  font-size: 72rpx;
}

.nickname {
  font-size: 36rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 12rpx;
}

.phone {
  font-size: 30rpx;
  color: $uni-text-color-grey;
  margin-bottom: 32rpx;
}

.house-card {
  background-color: #FAF8F5;
  border-radius: 16rpx;
  padding: 28rpx 40rpx;
  display: flex;
  flex-direction: column;
  align-items: center;

  .house-label {
    font-size: 24rpx;
    color: $uni-text-color-grey;
    margin-bottom: 8rpx;
  }

  .house-addr {
    font-size: 30rpx;
    font-weight: 600;
    color: $uni-color-primary;
  }
}
</style>
