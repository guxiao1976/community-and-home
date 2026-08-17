<template>
  <view class="page">
    <view class="header">
      <!-- 加入已由 join-community 立即完成（已建 membership），此页为「选择下一步」分流 -->
      <text class="header-title">已加入 {{ pendingCommunity?.communityName || '小区' }}</text>
      <text v-if="pendingCommunity?.address" class="header-addr">{{ pendingCommunity.address }}</text>
      <text class="header-sub">请选择下一步</text>
    </view>

    <!-- 身份分流：业主填房号 / 其他身份认证去我的页 -->
    <view v-if="pendingCommunity" class="choice-list">
      <view class="choice-card" hover-class="choice-card--hover" @click="goResidence">
        <text class="choice-icon">🏠</text>
        <text class="choice-title">填写房号成为业主</text>
        <text class="choice-desc">注册房号，成为业主后您可以使用物业报销、失物发布、邻里互动聊天、邻里互助等业主功能</text>
      </view>
      <view class="choice-card" hover-class="choice-card--hover" @click="goOtherAuth">
        <text class="choice-icon">🪪</text>
        <text class="choice-title">其他身份认证</text>
        <text class="choice-desc">可以申请为网格员、社区管理员、物业管理员</text>
      </view>
    </view>

    <!-- 深链兜底：无待加入小区 -->
    <view v-else class="empty">
      <text class="empty-text">请先选择待加入的小区</text>
      <view class="back-btn" @click="goBack">返回选择小区</view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { readPendingJoin, type PendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';

const communityStore = useCommunityStore();

// 待加入小区来自 join-community.vue 选中后存入的 pending-join 唯一契约源
// SEE: [[frontend-cross-page-storage-contract]] — 跨页临时数据收敛到共享模块
const pendingCommunity = ref<PendingJoin | null>(null);

onMounted(() => {
  pendingCommunity.value = readPendingJoin();
});

// 业主路径：pending 小区随行（仍存于 pending-join 内存态），join-residence 页消费
function goResidence() {
  if (!pendingCommunity.value) return;
  uni.navigateTo({ url: '/pages/join-residence/join-residence' });
}

// 其他身份认证路径：写一次性 pendingCommunityId 供我的页申请角色，消费后清除
function goOtherAuth() {
  const pending = pendingCommunity.value;
  if (!pending) return;
  communityStore.setPendingCommunityId(pending.communityId);
  uni.switchTab({ url: '/pages/my/my' });
}

function goBack() {
  uni.navigateBack();
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; background: #FFFFFF; padding: 0 2rem; }

.header { padding: 1.875rem 0 1.5rem; text-align: center;
  .header-title { display: block; font-size: 1.375rem; font-weight: 700; color: $uni-text-color; margin-bottom: 0.25rem; }
  .header-addr { display: block; font-size: 0.75rem; color: $uni-text-color-grey; margin-top: 0.25rem; }
  .header-sub { display: block; font-size: 0.9375rem; font-weight: 600; color: $uni-color-primary; margin-top: 0.375rem; }
}

.choice-list { display: flex; flex-direction: column; gap: 1rem; }
.choice-card { background: #FAF8F5; border-radius: 0.75rem; padding: 1.25rem 1rem; box-shadow: $uni-shadow-base; display: flex; flex-direction: column; align-items: center; text-align: center;
  &--hover { background: rgba(184, 149, 106, 0.08); }
  .choice-icon { font-size: 2rem; margin-bottom: 0.625rem; }
  .choice-title { font-size: 1.0625rem; font-weight: 700; color: $uni-text-color; margin-bottom: 0.5rem; }
  .choice-desc { font-size: 0.8125rem; color: $uni-text-color-grey; line-height: 1.6; }
}

.empty { padding: 3.75rem 0; text-align: center;
  .empty-text { display: block; font-size: 0.875rem; color: $uni-text-color-placeholder; margin-bottom: 1.25rem; }
  .back-btn { display: inline-block; padding: 0.5rem 1.5rem; border-radius: 1.375rem; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 0.875rem; font-weight: 600; }
}
</style>
