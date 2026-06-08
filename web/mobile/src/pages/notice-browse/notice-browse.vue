<template>
  <view class="page">
    <!-- 导航栏 -->
    <view class="nav-bar">
      <view class="nav-back" @click="goBack">← 返回</view>
      <text class="nav-title">通知公告</text>
      <text class="nav-count">{{ currentIndex + 1 }} / {{ notices.length }}</text>
    </view>

    <!-- 加载中 -->
    <view v-if="loading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 无通知 -->
    <view v-else-if="notices.length === 0" class="empty-wrap">
      <text class="empty-icon">📭</text>
      <text class="empty-text">暂无通知公告</text>
    </view>

    <!-- 通知内容 -->
    <view v-else class="notice-wrap">
      <view class="notice-card">
        <!-- 标题 -->
        <text class="notice-title">{{ currentNotice.title }}</text>

        <!-- 元信息 -->
        <view class="notice-meta">
          <view
            class="role-pill"
            :style="{ backgroundColor: roleColor + '18', borderColor: roleColor }"
          >
            <text class="role-text" :style="{ color: roleColor }">{{ roleName }}</text>
          </view>
          <text class="meta-time">{{ formatTime(currentNotice.publishedAt || currentNotice.createdAt) }}</text>
          <text class="meta-publisher">{{ currentNotice.publisher }}</text>
        </view>

        <!-- 分隔线 -->
        <view class="divider" />

        <!-- 正文 -->
        <view class="notice-content">
          <rich-text :nodes="currentNotice.content" />
        </view>
      </view>

      <!-- 翻页按钮 -->
      <view class="nav-btns">
        <view
          class="nav-btn"
          :class="{ 'nav-btn--disabled': currentIndex === 0 }"
          @click="prevNotice"
        >
          <text>← 上一条</text>
        </view>
        <view
          class="nav-btn"
          :class="{ 'nav-btn--disabled': currentIndex >= notices.length - 1 }"
          @click="nextNotice"
        >
          <text>下一条 →</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { getNoticeList, getNoticeRoleName, getNoticeRoleColor, type Notice } from '@/api/community';
import { useCommunityStore } from '@/stores/community';

const communityStore = useCommunityStore();

const notices = ref<Notice[]>([]);
const currentIndex = ref(0);
const loading = ref(true);

const currentNotice = computed(() => notices.value[currentIndex.value]);
const roleName = computed(() => getNoticeRoleName(currentNotice.value?.role || 0));
const roleColor = computed(() => getNoticeRoleColor(currentNotice.value?.role || 0));

function goBack() {
  uni.navigateBack();
}

function prevNotice() {
  if (currentIndex.value > 0) {
    currentIndex.value--;
  }
}

function nextNotice() {
  if (currentIndex.value < notices.value.length - 1) {
    currentIndex.value++;
  }
}

function formatTime(ts: number): string {
  if (!ts) return '';
  const d = new Date(ts * 1000);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

onMounted(async () => {
  const cid = communityStore.currentCommunityId;
  if (!cid) { loading.value = false; return; }
  try {
    const result = await getNoticeList(cid, 1, 50);
    // Filter recent 3 months
    const threeMonthsAgo = Math.floor(Date.now() / 1000) - 90 * 24 * 3600;
    notices.value = (result.notices || []).filter(
      (n: Notice) => (n.publishedAt || n.createdAt) >= threeMonthsAgo,
    );
  } catch {
    notices.value = [];
  } finally {
    loading.value = false;
  }
});
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
}

.nav-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #3D3226;
  margin-left: 16rpx;
  flex: 1;
}

.nav-count {
  font-size: 24rpx;
  color: #A6988A;
}

.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: #A6988A;
}

.empty-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 200rpx 0;
}

.empty-icon {
  font-size: 72rpx;
  margin-bottom: 16rpx;
}

.empty-text {
  font-size: 28rpx;
  color: #A6988A;
}

.notice-wrap {
  padding-top: 24rpx;
}

.notice-card {
  background: #FAF8F5;
  border-radius: 16rpx;
  padding: 32rpx;
  min-height: 400rpx;
}

.notice-title {
  font-size: 36rpx;
  font-weight: 700;
  color: #3D3226;
  display: block;
  line-height: 1.4;
  margin-bottom: 20rpx;
}

.notice-meta {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
  flex-wrap: wrap;
}

.role-pill {
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
  border-width: 1rpx;
  border-style: solid;
}

.role-text {
  font-size: 22rpx;
  font-weight: 600;
}

.meta-time {
  font-size: 24rpx;
  color: #A6988A;
}

.meta-publisher {
  font-size: 24rpx;
  color: #A6988A;
  margin-left: auto;
}

.divider {
  height: 1rpx;
  background: #E8DCCF;
  margin-bottom: 24rpx;
}

.notice-content {
  font-size: 28rpx;
  color: #3D3226;
  line-height: 1.8;
}

.nav-btns {
  display: flex;
  gap: 16rpx;
  margin-top: 32rpx;
  padding-bottom: 40rpx;
}

.nav-btn {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  background: #FAF8F5;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #3D3226;
  font-weight: 500;

  &--disabled {
    opacity: 0.3;
  }
}
</style>
