<template>
  <view class="page">
    <!-- 导航栏 -->
    <view class="nav-bar">
      <view class="nav-back" @click="goBack">← 返回</view>
      <text class="nav-title">通知公告</text>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 加载失败（REQ-NTW-4 场景 6，禁止静默吞错） -->
    <view v-else-if="loadError" class="empty-wrap">
      <text class="empty-icon">⚠️</text>
      <text class="empty-text">加载失败，请稍后重试</text>
    </view>

    <!-- 30 天窗口内无通知（REQ-NTW-4 场景 7） -->
    <view v-else-if="notices.length === 0" class="empty-wrap">
      <text class="empty-icon">📭</text>
      <text class="empty-text">暂无通知公告</text>
    </view>

    <!-- 30 天卡片列表（REQ-NTW-4/5 视觉契约：role 色条 + role 标签 + 标题 + 时间，置顶优先 + published_at 倒序由后端保证） -->
    <view v-else class="browse-list">
      <view
        v-for="item in notices"
        :key="item.id"
        class="browse-card"
        @click="onCardClick(item.id)"
      >
        <view class="browse-card-bar" :style="{ backgroundColor: getNoticeRoleColor(item.role) }" />
        <view class="browse-card-body">
          <view class="browse-card-header">
            <text class="browse-title">{{ item.title }}</text>
            <view
              class="browse-role-pill"
              :style="{
                backgroundColor: getNoticeRoleColor(item.role) + '18',
                borderColor: getNoticeRoleColor(item.role),
              }"
            >
              <text class="browse-role-text" :style="{ color: getNoticeRoleColor(item.role) }">
                {{ getNoticeRoleName(item.role) }}
              </text>
            </view>
          </view>
          <view class="browse-time-row">
            <text class="browse-time-icon">🕐</text>
            <text class="browse-time">{{ formatTime(item.published_at || item.created_at) }}</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { getNoticeList, getNoticeRoleName, getNoticeRoleColor, type Notice } from '@/api/community';
import { useCommunityStore } from '@/stores/community';
import dayjs from 'dayjs';

const communityStore = useCommunityStore();

const notices = ref<Notice[]>([]);
const loading = ref(true);
const loadError = ref(false);

function goBack() {
  uni.navigateBack();
}

function onCardClick(id: string) {
  uni.navigateTo({ url: `/pages/notice-detail/notice-detail?id=${id}` });
}

function formatTime(ts: number): string {
  if (!ts) return '';
  return dayjs.unix(ts).format('YYYY-MM-DD HH:mm');
}

onMounted(async () => {
  const cid = communityStore.currentCommunityId;
  if (!cid) { loading.value = false; return; }
  try {
    // 30 天窗口后端强制（since_days=30），前端只传参不实现窗口过滤（REQ-NTW-2）；
    // 单请求 page_size=50 截断，total 反映窗口内全量（REQ-NTW-4 场景 3）
    // SEE: [[frontend-business-rule-hardcode]]
    const result = await getNoticeList(cid, 1, 50, 30);
    notices.value = result.notices || [];
  } catch (e) {
    // SEE: [[verify-api-before-calling]] — 禁止静默吞错，明确失败态
    console.error('[notice-browse] 通知列表加载失败', e);
    loadError.value = true;
  } finally {
    loading.value = false;
  }
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #FFFFFF;
  padding: 0 1rem;
}

.nav-bar {
  display: flex;
  align-items: center;
  padding: 0.875rem 0;
  border-bottom: 0.03125rem solid #F0EBE3;
}

.nav-back {
  font-size: 0.875rem;
  color: #B8956A;
}

.nav-title {
  font-size: 1rem;
  font-weight: 600;
  color: #3D3226;
  margin-left: 0.5rem;
  flex: 1;
}

.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 6.25rem 0;
}

.loading-text {
  font-size: 0.875rem;
  color: #A6988A;
}

.empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6.25rem 0;
}

.empty-icon {
  font-size: 2.25rem;
  margin-bottom: 0.5rem;
}

.empty-text {
  font-size: 0.875rem;
  color: #A6988A;
}

// ---- 30 天卡片列表（与首页通知卡片一致）----
.browse-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  padding-top: 0.75rem;
}

.browse-card {
  display: flex;
  background: #FAF8F5;
  border-radius: 0.5rem;
  overflow: hidden;
  box-shadow: 0 0.0625rem 0.25rem rgba(0, 0, 0, 0.04);
}

.browse-card-bar {
  width: 0.3125rem;
  flex-shrink: 0;
}

.browse-card-body {
  flex: 1;
  padding: 0.75rem;
}

.browse-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 0.3125rem;
}

.browse-title {
  font-size: 0.9375rem;
  font-weight: 500;
  color: #3D3226;
  flex: 1;
  margin-right: 0.5rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.browse-role-pill {
  border-radius: 0.625rem;
  padding: 0.125rem 0.5rem;
  border: 0.03125rem solid;
  flex-shrink: 0;
}

.browse-role-text {
  font-size: 0.625rem;
  font-weight: 500;
}

.browse-time-row {
  display: flex;
  align-items: center;
  gap: 0.125rem;
}

.browse-time-icon {
  font-size: 0.625rem;
}

.browse-time {
  font-size: 0.6875rem;
  color: #A6988A;
}
</style>
