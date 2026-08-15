<template>
  <view class="page">
    <template v-if="notice">
      <!-- Title -->
      <text class="detail-title">{{ notice.title }}</text>

      <!-- Meta: role tag + time + publisher -->
      <view class="detail-meta">
        <view
          class="detail-role-tag"
          :style="{ backgroundColor: getNoticeRoleColor(notice.role) }"
        >
          <text class="role-text">{{ getNoticeRoleName(notice.role) }}</text>
        </view>
        <text class="detail-time">{{ formatTime(notice.published_at || notice.created_at) }}</text>
        <text v-if="notice.publisher" class="detail-publisher">{{ notice.publisher }}</text>
      </view>

      <view class="divider" />

      <!-- Content (rich text) -->
      <view class="detail-content">
        <rich-text :nodes="notice.content"></rich-text>
      </view>

      <!-- Attachments -->
      <template v-if="notice.attachments && notice.attachments.length > 0">
        <view class="divider" />
        <view class="detail-attachments">
          <text class="attachments-title">附件</text>
          <view
            v-for="att in notice.attachments"
            :key="att.id"
            class="attachment-item"
            @click="onDownload(att.file_url, att.file_name)"
          >
            <text class="attachment-icon">📎</text>
            <view class="attachment-info">
              <text class="attachment-name">{{ att.file_name }}</text>
              <text class="attachment-size">{{ formatSize(att.file_size) }}</text>
            </view>
          </view>
        </view>
      </template>
    </template>

    <!-- Loading / Error -->
    <view v-else-if="loading" class="status-wrap">
      <text class="status-text">加载中...</text>
    </view>
    <view v-else class="status-wrap">
      <text class="status-text">通知不存在</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getNoticeDetail, getNoticeRoleName, getNoticeRoleColor } from '@/api/community';
import type { Notice } from '@/api/community';
import dayjs from 'dayjs';

const notice = ref<Notice | null>(null);
const loading = ref(true);

function formatTime(ts: number): string {
  if (!ts) return '';
  return dayjs.unix(ts).format('YYYY-MM-DD HH:mm');
}

function formatSize(bytes: number): string {
  if (!bytes || bytes <= 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  let i = 0;
  let size = bytes;
  while (size >= 1024 && i < units.length - 1) {
    size /= 1024;
    i++;
  }
  return `${size.toFixed(i > 0 ? 1 : 0)} ${units[i]}`;
}

function onDownload(url: string, _name: string) {
  // #ifdef H5
  window.open(url, '_blank');
  // #endif
  // #ifdef MP-WEIXIN
  uni.downloadFile({
    url,
    success: (res) => {
      if (res.statusCode === 200) {
        uni.openDocument({ filePath: res.tempFilePath });
      }
    },
  });
  // #endif
}

async function fetchDetail(id: string) {
  loading.value = true;
  try {
    const data = await getNoticeDetail(id);
    notice.value = data.notice || null;
  } catch {
    notice.value = null;
  } finally {
    loading.value = false;
  }
}

onLoad((query) => {
  const id = query?.id as string;
  if (id) {
    fetchDetail(id);
  } else {
    loading.value = false;
  }
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
  padding: 32rpx;
}

.detail-title {
  display: block;
  font-size: 36rpx;
  font-weight: 700;
  color: $uni-text-color;
  line-height: 1.5;
  margin-bottom: 24rpx;
}

.detail-meta {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 16rpx;
}

.detail-role-tag {
  border-radius: 8rpx;
  padding: 4rpx 14rpx;
}

.role-text {
  font-size: 24rpx;
  color: #FFFFFF;
}

.detail-time {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

.detail-publisher {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

.divider {
  height: 1rpx;
  background-color: $uni-border-color;
  margin: 24rpx 0;
}

.detail-content {
  font-size: 28rpx;
  color: $uni-text-color;
  line-height: 1.8;
  word-break: break-all;
}

.detail-attachments {
  margin-top: 8rpx;
}

.attachments-title {
  display: block;
  font-size: 28rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 16rpx;
}

.attachment-item {
  display: flex;
  align-items: center;
  padding: 16rpx 0;
  border-bottom: 1rpx solid $uni-border-color;
}

.attachment-icon {
  font-size: 36rpx;
  margin-right: 16rpx;
}

.attachment-info {
  display: flex;
  flex-direction: column;
}

.attachment-name {
  font-size: 26rpx;
  color: $uni-color-primary;
}

.attachment-size {
  font-size: 22rpx;
  color: $uni-text-color-placeholder;
  margin-top: 4rpx;
}

.status-wrap {
  display: flex;
  justify-content: center;
  padding: 120rpx 0;
}

.status-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}
</style>
