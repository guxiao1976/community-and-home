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

      <!-- Attachments（REQ-NDP-1：无附件不渲染附件区） -->
      <template v-if="notice.attachments && notice.attachments.length > 0">
        <view class="divider" />
        <view class="detail-attachments">
          <text class="attachments-title">附件</text>
          <view
            v-for="att in notice.attachments"
            :key="att.id"
            class="attachment-item"
            @click="onAttachmentClick(att)"
          >
            <text class="attachment-icon">{{ isImageAttachment(att.file_type) ? '🖼️' : '📎' }}</text>
            <view class="attachment-info">
              <text class="attachment-name">{{ att.file_name }}</text>
              <text class="attachment-size">{{ formatSize(att.file_size) }}</text>
            </view>
          </view>
        </view>
      </template>
    </template>

    <!-- Loading / Error（REQ-NDP-1 场景 3：失败明确提示，不静默空白页） -->
    <view v-else-if="loading" class="status-wrap">
      <text class="status-text">加载中...</text>
    </view>
    <view v-else-if="loadError" class="status-wrap">
      <text class="status-text">加载失败，请稍后重试</text>
    </view>
    <view v-else class="status-wrap">
      <text class="status-text">通知不存在</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import {
  getNoticeDetail,
  getNoticeRoleName,
  getNoticeRoleColor,
  isImageAttachment,
} from '@/api/community';
import type { Notice, NoticeAttachment } from '@/api/community';
import dayjs from 'dayjs';

const notice = ref<Notice | null>(null);
const loading = ref(true);
const loadError = ref(false);

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

/**
 * 附件点击按 file_type 白名单分发（REQ-NDP-2/3，同一判定口径 isImageAttachment）：
 * 图片 → uni.previewImage 全屏预览；其余（pdf/doc/docx 或缺失/无法识别）→
 * uni.downloadFile 成功后 uni.openDocument。
 * 主路径消费详情响应中 community-hub 服务端重生的 file_url（预签名 URL），
 * 不直连 file-service REST（该端点强制文件所有权，查看者非上传者被拒）。
 * SEE: [[frontend-business-rule-hardcode]] — 白名单与 file-service guard/magic.go 对齐
 */
function onAttachmentClick(att: NoticeAttachment) {
  if (!att.file_url) {
    // 逐附件降级（legacy 无重生可能）：图片 → 预览失败；文档 → 附件打开失败（REQ-NDP-2/3 场景 2）
    if (isImageAttachment(att.file_type)) {
      uni.showToast({ title: '预览失败', icon: 'none' });
    } else {
      uni.showToast({ title: '附件打开失败', icon: 'none' });
    }
    return;
  }
  if (isImageAttachment(att.file_type)) {
    // 图片 → 全屏预览；预览失败明确提示，不降级到文档打开器（REQ-NDP-2 场景 2）
    uni.previewImage({
      urls: [att.file_url],
      fail: () => uni.showToast({ title: '预览失败', icon: 'none' }),
    });
  } else {
    // 文档 → 下载后 openDocument；下载失败明确提示（REQ-NDP-3 场景 2）
    uni.downloadFile({
      url: att.file_url,
      success: (res: any) => {
        if (res?.statusCode === 200 && res.tempFilePath) {
          uni.openDocument({ filePath: res.tempFilePath, showMenu: true });
        } else {
          uni.showToast({ title: '附件打开失败', icon: 'none' });
        }
      },
      fail: () => uni.showToast({ title: '附件打开失败', icon: 'none' }),
    });
  }
}

async function fetchDetail(id: string) {
  loading.value = true;
  loadError.value = false;
  try {
    const data = await getNoticeDetail(id);
    notice.value = data.notice || null;
  } catch (e) {
    // 详情加载失败（API 失败/详情读整单失败）→ 明确失败态，禁止静默吞错
    console.error('[notice-detail] 详情加载失败', e);
    notice.value = null;
    loadError.value = true;
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
