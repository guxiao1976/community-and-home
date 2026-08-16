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
  // 访问者身份一律以已认证用户为准——禁止从 URL 取 viewer_id（可手工构造 → IDOR 越权看他人手机号/房屋号）。
  // 数据范围/脱敏决策在后端，前端只传自己的身份作上下文。
  // SEE: [[api-accessor-identity-from-url]]
  const viewerId = userStore.userId || undefined;
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
  padding: 6.25rem 0;

  .state-text {
    font-size: 0.875rem;
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
  padding: 3.125rem 1.25rem;
}

.avatar {
  width: 4.375rem;
  height: 4.375rem;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 0.09375rem solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 0.125rem 0.625rem rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.avatar-emoji {
  font-size: 2.25rem;
}

.nickname {
  font-size: 1.125rem;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 0.375rem;
}

.phone {
  font-size: 0.9375rem;
  color: $uni-text-color-grey;
  margin-bottom: 1rem;
}

.house-card {
  background-color: #FAF8F5;
  border-radius: 0.5rem;
  padding: 0.875rem 1.25rem;
  display: flex;
  flex-direction: column;
  align-items: center;

  .house-label {
    font-size: 0.75rem;
    color: $uni-text-color-grey;
    margin-bottom: 0.25rem;
  }

  .house-addr {
    font-size: 0.9375rem;
    font-weight: 600;
    color: $uni-color-primary;
  }
}
</style>
