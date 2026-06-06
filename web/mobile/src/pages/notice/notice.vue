<template>
  <view class="page">
    <!-- 小区切换器（含渐变头部） -->
    <CommunitySwitcher
      :communities="communityStore.communities"
      :model-value="communityStore.currentCommunityId"
      @switch="onCommunitySwitch"
    />

    <!-- 未加入小区提示 -->
    <view v-if="!communityStore.hasCommunities && !loading" class="no-community-hint">
      <text class="no-community-icon">📍</text>
      <text class="no-community-text">请先加入小区以查看更多内容</text>
      <text class="no-community-link" @click="goJoinCommunity">去加入 →</text>
    </view>

    <!-- 骨架屏 -->
    <template v-if="loading">
      <view class="skeleton-wrap">
        <!-- 通知骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title" />
          <view v-for="i in 3" :key="'sn'+i" class="skeleton-notice">
            <view class="skeleton-bar" />
            <view class="skeleton-text" />
          </view>
        </view>
        <!-- 联络骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title short" />
          <view class="skeleton-grid">
            <view v-for="i in 4" :key="'sc'+i" class="skeleton-contact" />
          </view>
        </view>
        <!-- 寻失骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title short" />
          <view class="skeleton-lf-row">
            <view v-for="i in 2" :key="'slf'+i" class="skeleton-lf-card" />
          </view>
        </view>
      </view>
    </template>

    <!-- 实际内容 -->
    <template v-else-if="communityStore.hasCommunities">
      <!-- 通知公告 -->
      <view class="section">
        <view class="section-header">
          <view class="section-header-left">
            <text class="section-icon">📢</text>
            <text class="section-title">通知公告</text>
          </view>
          <text class="section-more" @click="onMoreNotice">全部 →</text>
        </view>
        <view v-if="notices.length === 0" class="empty-state">
          <text class="empty-icon">📭</text>
          <text class="empty-text">暂无通知公告</text>
        </view>
        <view v-else class="notice-list">
          <view
            v-for="item in notices"
            :key="item.id"
            class="notice-card"
            @click="onNoticeClick(item.id)"
          >
            <view class="notice-bar" :style="{ backgroundColor: getNoticeRoleColor(item.role) }" />
            <view class="notice-body">
              <view class="notice-card-header">
                <text class="notice-title">{{ item.title }}</text>
                <view
                  class="notice-role-pill"
                  :style="{
                    backgroundColor: getNoticeRoleColor(item.role) + '18',
                    borderColor: getNoticeRoleColor(item.role),
                  }"
                >
                  <text class="role-text" :style="{ color: getNoticeRoleColor(item.role) }">
                    {{ getNoticeRoleName(item.role) }}
                  </text>
                </view>
              </view>
              <view class="notice-time-row">
                <text class="notice-time-icon">🕐</text>
                <text class="notice-time">{{ formatTime(item.publishedAt || item.createdAt) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- 便民联络 -->
      <view class="section">
        <view class="section-header">
          <view class="section-header-left">
            <text class="section-icon">📞</text>
            <text class="section-title">便民联络</text>
          </view>
        </view>
        <view v-if="contactGroups.length === 0" class="empty-state">
          <text class="empty-icon">📞</text>
          <text class="empty-text">暂无联络信息</text>
        </view>
        <view v-else class="contact-grid">
          <view
            v-for="(group, gIdx) in contactGroups"
            :key="gIdx"
            class="contact-card"
            @click="onCallPhone(group.phone)"
          >
            <text class="contact-icon">{{ group.icon }}</text>
            <view class="contact-body">
              <text class="contact-category">{{ group.categoryName }}</text>
              <text class="contact-phone">{{ group.phone }}</text>
            </view>
          </view>
        </view>
      </view>

      <!-- 广告位 ×2（便民联络下方） -->
      <view
        class="ad-banner"
        :style="{ backgroundImage: 'url(https://images.unsplash.com/photo-1542838132-92c53300491e?w=800&h=200&fit=crop)' }"
      >
        <view class="ad-overlay">
          <view class="ad-text">
            <text class="ad-label">限时特惠</text>
            <text class="ad-title">社区团购 · 生鲜直达 🥬</text>
            <text class="ad-desc">今日下单 明日送达 满50减10</text>
          </view>
          <view class="ad-btn" style="background-color: #FF6B35;" @click.stop="onAdClick('groupbuy')">
            <text>去看看</text>
          </view>
        </view>
      </view>
      <view
        class="ad-banner ad-banner--spaced"
        :style="{ backgroundImage: 'url(https://images.unsplash.com/photo-1581578731548-c64695cc6952?w=800&h=200&fit=crop)' }"
      >
        <view class="ad-overlay">
          <view class="ad-text">
            <text class="ad-label">新用户专享</text>
            <text class="ad-title">家政保洁 · 首单立减 🧹</text>
            <text class="ad-desc">专业认证保洁师 2小时69元起</text>
          </view>
          <view class="ad-btn" style="background-color: #0EA5E9;" @click.stop="onAdClick('housekeeping')">
            <text>立即预约</text>
          </view>
        </view>
      </view>

      <!-- 寻失互助 -->
      <view class="section">
        <view class="section-header">
          <view class="section-header-left">
            <text class="section-icon">🔍</text>
            <text class="section-title">寻失互助</text>
          </view>
        </view>
        <view v-if="lostFoundItems.length === 0" class="empty-state">
          <text class="empty-icon">🔍</text>
          <text class="empty-text">暂无寻失信息</text>
        </view>
        <view v-else class="lost-found-wrap">
          <view class="lost-found-list">
            <view
              v-for="item in lostFoundItems"
              :key="item.id"
              class="lost-found-card"
              @click="onLostFoundClick(item.id)"
            >
              <view class="lost-found-image-wrap">
                <image
                  v-if="item.imageUrls && item.imageUrls.length > 0 && !imageErrors.has(item.id)"
                  :src="item.imageUrls[0]"
                  class="lost-found-image"
                  mode="aspectFill"
                  @error="onImageError($event, item.id)"
                />
                <text v-else class="lost-found-placeholder">🖼️</text>
                <view
                  class="lost-found-tag"
                  :style="{ backgroundColor: item.type === 1 ? '#B8956A' : '#8DAF7E' }"
                >
                  <text class="lost-found-tag-text">{{ getLostFoundTypeName(item.type) }}</text>
                </view>
              </view>
              <view class="lost-found-body">
                <text class="lost-found-title">{{ item.title }}</text>
                <text class="lost-found-time">{{ formatTime(item.createdAt) }}</text>
              </view>
            </view>
          </view>
          <text class="section-more section-more--center">全部 →</text>
        </view>
      </view>

      <!-- 广告位 ×1（寻失互助下方） -->
      <view
        class="ad-banner ad-banner--spaced"
        :style="{ backgroundImage: 'url(https://images.unsplash.com/photo-1523050854058-8df90910b48a?w=800&h=200&fit=crop)' }"
      >
        <view class="ad-overlay">
          <view class="ad-text">
            <text class="ad-label">免费试听</text>
            <text class="ad-title">社区课堂 · 兴趣培养 📚</text>
            <text class="ad-desc">书法 绘画 舞蹈 家门口的兴趣班</text>
          </view>
          <view class="ad-btn" style="background-color: #8B5CF6;" @click.stop="onAdClick('classroom')">
            <text>免费试听</text>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { onPullDownRefresh } from '@dcloudio/uni-app';
import { useCommunityStore } from '@/stores/community';
import {
  getNoticeList,
  getContacts,
  getLostFoundList,
  getNoticeRoleName,
  getNoticeRoleColor,
  getContactCategoryName,
  getContactCategoryIcon,
  getLostFoundTypeName,
} from '@/api/community';
import type { Notice, Contact, LostFoundItem } from '@/api/community';
import CommunitySwitcher from '@/components/community-switcher.vue';
import dayjs from 'dayjs';

const communityStore = useCommunityStore();

// ---- Loading ----
const loading = ref(true);

// ---- Notice State ----
const notices = ref<Notice[]>([]);

// ---- Contact State ----
const contacts = ref<Contact[]>([]);

interface ContactGroup {
  category: number;
  categoryName: string;
  icon: string;
  phone: string;
}

const contactGroups = computed<ContactGroup[]>(() => {
  const map = new Map<number, Contact[]>();
  for (const c of contacts.value) {
    const list = map.get(c.category) || [];
    list.push(c);
    map.set(c.category, list);
  }
  const groups: ContactGroup[] = [];
  for (const [cat, items] of map) {
    for (const item of items) {
      groups.push({
        category: cat,
        categoryName: getContactCategoryName(cat),
        icon: getContactCategoryIcon(cat),
        phone: item.phone,
      });
    }
  }
  return groups;
});

// ---- Lost & Found State ----
const lostFoundItems = ref<LostFoundItem[]>([]);
const imageErrors = ref<Set<string>>(new Set());

// ---- Actions ----
function onMoreNotice() {
  uni.showToast({ title: '通知列表开发中...', icon: 'none', duration: 2000 });
}

function onNoticeClick(id: string) {
  uni.navigateTo({ url: `/pages/notice-detail/notice-detail?id=${id}` });
}

function onCallPhone(phone: string) {
  uni.makePhoneCall({ phoneNumber: phone });
}

function onImageError(_event: Event, itemId: string) {
  imageErrors.value.add(itemId);
}

function onLostFoundClick(_id: string) {
  uni.showToast({ title: '寻失详情开发中...', icon: 'none', duration: 2000 });
}

function onAdClick(_type: string) {
  // 广告点击预留，暂不跳转
}

function onCommunitySwitch(id: string) {
  communityStore.switchCommunity(id);
}

function goJoinCommunity() {
  uni.navigateTo({ url: '/pages/join-community/join-community' });
}

function formatTime(ts: number): string {
  if (!ts) return '';
  return dayjs.unix(ts).format('YYYY-MM-DD HH:mm');
}

// ---- Fetch Data ----
async function fetchNotices() {
  const cid = communityStore.currentCommunityId;
  if (!cid) return;
  try {
    const data = await getNoticeList(cid);
    notices.value = data.notices || [];
  } catch { /* silent */ }
}

async function fetchContacts() {
  const cid = communityStore.currentCommunityId;
  if (!cid) return;
  try {
    const data = await getContacts(cid);
    contacts.value = data.contacts || [];
  } catch { /* silent */ }
}

async function fetchLostFound() {
  const cid = communityStore.currentCommunityId;
  if (!cid) return;
  try {
    const data = await getLostFoundList(cid);
    lostFoundItems.value = data.items || [];
  } catch { /* silent */ }
}

async function loadAll() {
  loading.value = true;
  await Promise.all([fetchNotices(), fetchContacts(), fetchLostFound()]);
  loading.value = false;
}

// ---- Lifecycle ----
watch(() => communityStore.currentCommunityId, (newVal, oldVal) => {
  if (newVal && newVal !== oldVal) {
    loadAll();
  }
});

onMounted(async () => {
  await communityStore.loadMemberships();
  loadAll();
});

// ---- Pull to Refresh ----
onPullDownRefresh(async () => {
  await loadAll();
  uni.stopPullDownRefresh();
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: $uni-bg-color;
  padding-bottom: calc(100rpx + env(safe-area-inset-bottom));
}

// ---- No Community Hint ----
.no-community-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 32rpx;
}

.no-community-icon {
  font-size: 80rpx;
  margin-bottom: 24rpx;
  opacity: 0.6;
}

.no-community-text {
  font-size: 28rpx;
  color: $uni-text-color-grey;
  margin-bottom: 24rpx;
}

.no-community-link {
  font-size: 28rpx;
  color: $uni-color-primary;
  font-weight: 600;
  padding: 16rpx 48rpx;
  border-radius: 48rpx;
  background: rgba($uni-color-primary, 0.08);
}

// ---- Skeleton ----
.skeleton-wrap {
  padding: 0 32rpx;
}

.skeleton-section {
  margin-bottom: 32rpx;
}

.skeleton-title {
  width: 160rpx;
  height: 34rpx;
  border-radius: 8rpx;
  background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  margin-bottom: 20rpx;

  &.short {
    width: 120rpx;
  }
}

.skeleton-notice {
  display: flex;
  align-items: center;
  background: $uni-bg-color-card;
  border-radius: 12rpx;
  padding: 24rpx;
  margin-bottom: 12rpx;

  .skeleton-bar {
    width: 8rpx;
    height: 56rpx;
    border-radius: 4rpx;
    background: linear-gradient(180deg, $uni-bg-color-grey 25%, #E8E0D5 50%, $uni-bg-color-grey 75%);
    background-size: 100% 200%;
    animation: shimmer 1.5s infinite;
    margin-right: 20rpx;
    flex-shrink: 0;
  }

  .skeleton-text {
    flex: 1;
    height: 28rpx;
    border-radius: 6rpx;
    background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }
}

.skeleton-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
}

.skeleton-contact {
  height: 72rpx;
  border-radius: 10rpx;
  background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-lf-row {
  display: flex;
  gap: 20rpx;
  justify-content: center;
}

.skeleton-lf-card {
  width: 300rpx;
  height: 200rpx;
  border-radius: 14rpx;
  background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

@keyframes shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

// ---- Section Shared ----
.section {
  padding: 0 32rpx;
  margin-bottom: 36rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.section-header-left {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.section-icon {
  font-size: 36rpx;
}

.section-title {
  font-size: 34rpx;
  font-weight: 600;
  color: $uni-text-color;
}

.section-more {
  font-size: 24rpx;
  color: $uni-color-primary;

  &--center {
    display: block;
    text-align: center;
    margin-top: 12rpx;
  }
}

// ---- Empty State ----
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60rpx 0;
}

.empty-icon {
  font-size: 72rpx;
  margin-bottom: 16rpx;
  opacity: 0.6;
}

.empty-text {
  font-size: 26rpx;
  color: $uni-text-color-placeholder;
}

// ---- Notice Cards ----
.notice-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.notice-card {
  display: flex;
  background-color: $uni-bg-color-card;
  border-radius: 12rpx;
  overflow: hidden;
  box-shadow: $uni-shadow-sm;
}

.notice-bar {
  width: 10rpx;
  flex-shrink: 0;
}

.notice-body {
  flex: 1;
  padding: 22rpx 24rpx 18rpx;
}

.notice-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8rpx;
}

.notice-title {
  font-size: 28rpx;
  font-weight: 500;
  color: $uni-text-color;
  flex: 1;
  margin-right: 16rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.notice-role-pill {
  border-radius: 20rpx;
  padding: 4rpx 16rpx;
  border: 1px solid;
  flex-shrink: 0;
}

.role-text {
  font-size: 20rpx;
  font-weight: 500;
}

.notice-time-row {
  display: flex;
  align-items: center;
  gap: 4rpx;
}

.notice-time-icon {
  font-size: 20rpx;
}

.notice-time {
  font-size: 22rpx;
  color: $uni-text-color-grey;
}

// ---- Contact Grid ----
.contact-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12rpx;
}

.contact-card {
  display: flex;
  align-items: center;
  background-color: $uni-bg-color-card;
  border-radius: 10rpx;
  padding: 14rpx 16rpx;
  box-shadow: $uni-shadow-sm;
}

.contact-icon {
  font-size: 36rpx;
  margin-right: 12rpx;
  flex-shrink: 0;
}

.contact-body {
  flex: 1;
  min-width: 0;
}

.contact-category {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-bottom: 2rpx;
}

.contact-phone {
  display: block;
  font-size: 28rpx;
  font-weight: 500;
  color: #5D5348;
}

// ---- Ad Banner ----
.ad-banner {
  height: 180rpx;
  background-size: cover;
  background-position: center;
  position: relative;
  overflow: hidden;

  &--spaced {
    margin-top: 24rpx;
    margin-bottom: 36rpx;
  }
}

.ad-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(0,0,0,0.35) 0%, rgba(0,0,0,0.05) 100%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 28rpx;
}

.ad-text {
  display: flex;
  flex-direction: column;
}

.ad-label {
  font-size: 20rpx;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 2rpx;
}

.ad-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #fff;
  margin-bottom: 2rpx;
}

.ad-desc {
  font-size: 22rpx;
  color: rgba(255, 255, 255, 0.85);
}

.ad-btn {
  padding: 10rpx 24rpx;
  border-radius: 32rpx;
  flex-shrink: 0;

  text {
    color: #fff;
    font-size: 24rpx;
    font-weight: 600;
  }
}

// ---- Lost & Found ----
.lost-found-wrap {
  display: flex;
  flex-direction: column;
}

.lost-found-list {
  display: flex;
  justify-content: center;
  gap: 24rpx;
}

.lost-found-card {
  width: 300rpx;
  flex-shrink: 0;
  background-color: $uni-bg-color-card;
  border-radius: 14rpx;
  overflow: hidden;
  box-shadow: $uni-shadow-sm;
}

.lost-found-image-wrap {
  width: 100%;
  height: 200rpx;
  background: linear-gradient(135deg, $uni-bg-color-grey, $uni-border-color);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  position: relative;
}

.lost-found-image {
  width: 100%;
  height: 100%;
}

.lost-found-placeholder {
  font-size: 80rpx;
  opacity: 0.4;
}

.lost-found-tag {
  position: absolute;
  top: 12rpx;
  left: 12rpx;
  border-radius: 8rpx;
  padding: 4rpx 14rpx;
}

.lost-found-tag-text {
  font-size: 20rpx;
  color: #fff;
}

.lost-found-body {
  padding: 14rpx 16rpx 18rpx;
}

.lost-found-title {
  display: block;
  font-size: 26rpx;
  font-weight: 500;
  color: $uni-text-color;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 4rpx;
}

.lost-found-time {
  font-size: 20rpx;
  color: $uni-text-color-grey;
}
</style>
