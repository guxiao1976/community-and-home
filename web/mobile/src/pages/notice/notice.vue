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
    <!-- REQ-HL-4 固定垂直全序：通知 → 4功能入口 → 邻里互助占位 → 寻失互助 → 底部广告集中区 -->
    <template v-else-if="communityStore.hasCommunities">
      <!-- ① 通知公告（标题栏 + 30 天卡片列表） -->
      <view class="section">
        <view class="section-header notice-header">
          <view class="section-header-left">
            <text class="section-icon">📢</text>
            <text class="section-title">通知公告</text>
          </view>
          <text class="section-more" @click="onMoreNotice">更多</text>
        </view>
        <!-- 通知为空时不渲染列表与空态块，仅保留标题栏，下方内容自然上移 -->
        <view v-if="notices.length > 0" class="notice-list">
          <view
            v-for="item in notices"
            :key="item.id"
            class="notice-card"
            @click="onNoticeClick(item.id)"
          >
            <view class="notice-bar" :style="{ backgroundColor: getNoticeRoleColor(item.role) }" />
            <view class="notice-body">
              <!-- 标题行：全文显示、自然换行（不截断、无省略号） -->
              <text class="notice-title">{{ item.title }}</text>
              <!-- 元信息行：发布单位 + 发布日期（YYYY-MM-DD） -->
              <view class="notice-meta">
                <text class="notice-publisher">{{ getPublisherName(item) }}</text>
                <text class="notice-date">{{ formatPublishDate(item.published_at || item.created_at) }}</text>
              </view>
            </view>
          </view>
        </view>
      </view>

      <!-- ② 4 功能图标入口（REQ-FE-1/2/3） -->
      <view class="section">
        <view class="func-entries">
          <view
            v-for="entry in FUNCTION_ENTRIES"
            :key="entry.key"
            class="func-entry"
            @click="onFuncEntry(entry)"
          >
            <text class="func-entry-icon">{{ entry.icon }}</text>
            <text class="func-entry-label">{{ entry.label }}</text>
          </view>
        </view>
      </view>

      <!-- ③ 邻里互助占位区块（REQ-HL-1，本期无后端/无页面，点击不导航、不伪造数据） -->
      <view class="section">
        <view class="section-header">
          <view class="section-header-left">
            <text class="section-icon">🤝</text>
            <text class="section-title">邻里互助</text>
          </view>
        </view>
        <view class="empty-state">
          <text class="empty-icon">🤝</text>
          <text class="empty-text">互助功能开发中</text>
        </view>
      </view>

      <!-- ④ 寻失互助（REQ-HL-2，样式与数据保持现状） -->
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
                  v-if="item.image_urls && item.image_urls.length > 0 && !imageErrors.has(item.id)"
                  :src="item.image_urls[0]"
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
                <text class="lost-found-time">{{ formatTime(item.created_at) }}</text>
              </view>
            </view>
          </view>
          <text class="section-more section-more--center">全部 →</text>
        </view>
      </view>

      <!-- ⑤ 底部广告集中区（REQ-HL-3：3 个广告垂直堆叠，内容保留硬编码，点击预留不跳转） -->
      <view class="ad-section">
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
          class="ad-banner"
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
        <view
          class="ad-banner"
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
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { onPullDownRefresh } from '@dcloudio/uni-app';
import { useCommunityStore } from '@/stores/community';
import {
  getNoticeList,
  getLostFoundList,
  getNoticeRoleName,
  getNoticeRoleColor,
  getLostFoundTypeName,
} from '@/api/community';
import type { Notice, LostFoundItem } from '@/api/community';
import CommunitySwitcher from '@/components/community-switcher.vue';
import dayjs from 'dayjs';

const communityStore = useCommunityStore();

// ---- 首载守卫（REQ-DBL-1）：初始进入时同批接口（通知+寻失）只拉一遍 ----
// loadMemberships 内 getAppState 服务端权威覆写 currentCommunityId 会触发 watch 一次；
// 该标志默认 false，watch 在 `!membershipsResolved` 时直接忽略（守卫已就绪前的变更不加载）。
// 关键时序（评审钉死）：标志只能在 notice.vue 的 onMounted 中 `await loadMemberships()`
// 之后置 true —— 若放进 loadMemberships 内部（含 finally），getAppState 覆写触发的 watch
// 会在标志已 true 时执行 → 双重加载依旧（正是本变更要修的缺陷）。
// SEE: [[frontend-business-rule-hardcode]] — 当前小区权威在后端（getAppState），前端仅消费
let membershipsResolved = false;

// ---- Loading ----
const loading = ref(true);

// ---- Notice State ----
const notices = ref<Notice[]>([]);

// ---- Lost & Found State ----
const lostFoundItems = ref<LostFoundItem[]>([]);
const imageErrors = ref<Set<string>>(new Set());

// ---- 4 功能图标入口（REQ-FE-1/2/3）----
interface FuncEntry {
  key: string;
  label: string;
  icon: string;
  // target 存在 → 做实跳页；不存在 → 占位 toast「功能开发中」不跳转
  target?: string;
}

const FUNCTION_ENTRIES: FuncEntry[] = [
  { key: 'contact', label: '便民联络', icon: '📞', target: '/pages/contact-list/contact-list' },
  { key: 'repair', label: '物业报修', icon: '🔧' },
  { key: 'secondhand', label: '二手闲置', icon: '📦' },
  { key: 'rent', label: '租房卖房', icon: '🏠' },
];

function onFuncEntry(entry: FuncEntry) {
  if (entry.target) {
    uni.navigateTo({ url: entry.target });
  } else {
    uni.showToast({ title: '功能开发中', icon: 'none' });
  }
}

// ---- 通知两行布局（标题全文 + 元信息行）----
// 发布日期：YYYY-MM-DD（年月日）。published_at=0 时回退 created_at。
function formatPublishDate(ts: number): string {
  if (!ts) return '';
  return dayjs.unix(ts).format('YYYY-MM-DD');
}

// 发布单位：item.publisher 非空则用它，否则回退 getNoticeRoleName(item.role)。
function getPublisherName(item: Notice): string {
  return item.publisher && item.publisher.trim() ? item.publisher : getNoticeRoleName(item.role);
}

// ---- Actions ----
function onMoreNotice() {
  // 空态也允许进浏览页（移除 notices.length===0 拦截）
  uni.navigateTo({ url: '/pages/notice-browse/notice-browse' });
}

function onNoticeClick(id: string) {
  uni.navigateTo({ url: `/pages/notice-detail/notice-detail?id=${id}` });
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

async function onCommunitySwitch(id: string) {
  try {
    await communityStore.switchCommunity(id);
  } catch (e: unknown) {
    const err = e as { code?: number };
    // 10015 目标小区不在数据范围：明确提示，当前小区保持不变
    if (err?.code === 10015) {
      uni.showToast({ title: '目标小区不在你的数据范围', icon: 'none', duration: 2500 });
      return;
    }
    // 非 10015 不再静默：console.error + 通用 toast
    console.error('[notice] 切换小区失败', e);
    uni.showToast({ title: '切换小区失败', icon: 'none' });
  }
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
    // since_days=30（30 天窗口后端强制，前端只传参）+ page_size=3（首页 ≤3 条）
    // SEE: [[frontend-business-rule-hardcode]] — 窗口业务逻辑在后端
    const data = await getNoticeList(cid, 1, 3, 30);
    notices.value = data.notices || [];
  } catch (e) {
    // SEE: [[verify-api-before-calling]] — 禁止静默吞错，失败需 console.error + 区块提示
    console.error('[notice] 通知加载失败', e);
    uni.showToast({ title: '通知加载失败', icon: 'none' });
  }
}

async function fetchLostFound() {
  const cid = communityStore.currentCommunityId;
  if (!cid) return;
  try {
    const data = await getLostFoundList(cid);
    lostFoundItems.value = data.items || [];
  } catch (e) {
    console.error('[notice] 寻失加载失败', e);
    uni.showToast({ title: '寻失加载失败', icon: 'none' });
  }
}

async function loadAll() {
  loading.value = true;
  await Promise.all([fetchNotices(), fetchLostFound()]);
  loading.value = false;
}

// ---- Lifecycle ----
watch(() => communityStore.currentCommunityId, (newVal, oldVal) => {
  // 首载守卫：loadMemberships 内 getAppState 覆写 currentCommunityId 的这次变更在
  // 标志置位前触发，被忽略；用户手动切换小区时标志已 true，正常触发单次 loadAll。
  if (!membershipsResolved) return;
  if (newVal && newVal !== oldVal) {
    loadAll();
  }
});

onMounted(async () => {
  await communityStore.loadMemberships();
  // 守卫解除必须在 loadMemberships 完成后（见模块顶部注释），随后显式单次 loadAll。
  // loadMemberships 整体失败时 communities 被清空（hasCommunities=false）→ 不发请求，
  // 不以陈旧 currentCommunityId 拉数据（REQ-DBL-1 降级）；守卫已就绪，后续切换仍可触发。
  membershipsResolved = true;
  if (communityStore.hasCommunities) {
    loadAll();
  } else {
    // 无小区：不发起数据加载（不以陈旧 cid 发请求，REQ-DBL-1 降级），
    // 直接结束骨架屏，展示「请先加入小区」空态引导（既有行为不回归）
    loading.value = false;
  }
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
  padding-bottom: calc(3.125rem + env(safe-area-inset-bottom));
}

// ---- No Community Hint ----
.no-community-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3.75rem 1rem;
}

.no-community-icon {
  font-size: 2.5rem;
  margin-bottom: 0.75rem;
  opacity: 0.6;
}

.no-community-text {
  font-size: 0.875rem;
  color: $uni-text-color-grey;
  margin-bottom: 0.75rem;
}

.no-community-link {
  font-size: 0.875rem;
  color: $uni-color-primary;
  font-weight: 600;
  padding: 0.5rem 1.5rem;
  border-radius: 1.5rem;
  background: rgba($uni-color-primary, 0.08);
}

// ---- Skeleton ----
.skeleton-wrap {
  padding: 0 1rem;
}

.skeleton-section {
  margin-bottom: 1rem;
}

.skeleton-title {
  width: 5rem;
  height: 1.0625rem;
  border-radius: 0.25rem;
  background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
  margin-bottom: 0.625rem;

  &.short {
    width: 3.75rem;
  }
}

.skeleton-notice {
  display: flex;
  align-items: center;
  background: $uni-bg-color-card;
  border-radius: 0.375rem;
  padding: 0.75rem;
  margin-bottom: 0.375rem;

  .skeleton-bar {
    width: 0.25rem;
    height: 1.75rem;
    border-radius: 0.125rem;
    background: linear-gradient(180deg, $uni-bg-color-grey 25%, #E8E0D5 50%, $uni-bg-color-grey 75%);
    background-size: 100% 200%;
    animation: shimmer 1.5s infinite;
    margin-right: 0.625rem;
    flex-shrink: 0;
  }

  .skeleton-text {
    flex: 1;
    height: 0.875rem;
    border-radius: 0.1875rem;
    background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
    background-size: 200% 100%;
    animation: shimmer 1.5s infinite;
  }
}

.skeleton-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.375rem;
}

.skeleton-contact {
  height: 2.25rem;
  border-radius: 0.3125rem;
  background: linear-gradient(90deg, $uni-bg-color-grey 25%, $uni-bg-color-card 50%, $uni-bg-color-grey 75%);
  background-size: 200% 100%;
  animation: shimmer 1.5s infinite;
}

.skeleton-lf-row {
  display: flex;
  gap: 0.625rem;
  justify-content: center;
}

.skeleton-lf-card {
  width: 9.375rem;
  height: 6.25rem;
  border-radius: 0.4375rem;
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
  padding: 0 1rem;
  margin-bottom: 1.125rem;
}

// ---- Notice Header（标题栏，任务 4）----
.notice-header {
  margin-bottom: 0.5rem;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.section-header-left {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.section-icon {
  font-size: 1.125rem;
}

.section-title {
  font-size: 1.0625rem;
  font-weight: 600;
  color: $uni-text-color;
}

.section-more {
  font-size: 0.75rem;
  color: $uni-color-primary;

  &--center {
    display: block;
    text-align: center;
    margin-top: 0.375rem;
  }
}

// ---- Empty State ----
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 1.875rem 0;
}

.empty-icon {
  font-size: 2.25rem;
  margin-bottom: 0.5rem;
  opacity: 0.6;
}

.empty-text {
  font-size: 0.8125rem;
  color: $uni-text-color-placeholder;
}

// ---- Notice Cards（两行布局：标题全文 + 元信息行）----
// 简洁样式：透明底（与页面底色一致，无卡片背景/阴影/圆角），细分隔线区分行
.notice-list {
  display: flex;
  flex-direction: column;
}

.notice-card {
  display: flex;
  padding: 0.6875rem 0;
  border-bottom: 0.0625rem solid $uni-border-color;

  &:last-child {
    border-bottom: none;
  }
}

.notice-bar {
  width: 0.3125rem;
  flex-shrink: 0;
  border-radius: 0.125rem;
}

.notice-body {
  flex: 1;
  padding: 0 0 0 0.625rem;
  min-width: 0;
}

// 标题行：全文显示、自然换行（white-space normal，无 nowrap / JS 截断）
.notice-title {
  display: block;
  font-size: 0.9375rem;
  font-weight: 500;
  line-height: 1.5;
  color: $uni-text-color;
  white-space: normal;
  word-break: break-all;
}

// 元信息行：发布单位 + 发布日期（YYYY-MM-DD）
.notice-meta {
  display: flex;
  align-items: center;
  margin-top: 0.25rem;
}

.notice-publisher {
  font-size: 0.6875rem;
  color: $uni-text-color-grey;
}

.notice-date {
  font-size: 0.6875rem;
  color: $uni-text-color-placeholder;
  margin-left: 0.5rem;
}

// ---- 4 Function Entries (REQ-FE-1) ----
.func-entries {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 0.5rem;
}

.func-entry {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  background-color: $uni-bg-color-card;
  border-radius: 0.5rem;
  padding: 0.75rem 0.25rem;
  box-shadow: $uni-shadow-sm;
}

.func-entry-icon {
  font-size: 1.375rem;
  margin-bottom: 0.3125rem;
}

.func-entry-label {
  font-size: 0.75rem;
  color: $uni-text-color;
  font-weight: 500;
}

// ---- Ad Banner（底部集中区，REQ-HL-3）----
.ad-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  padding: 0 1rem 1rem;
}

.ad-banner {
  height: 5.625rem;
  background-size: cover;
  background-position: center;
  position: relative;
  overflow: hidden;
  border-radius: 0.375rem;
}

.ad-overlay {
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, rgba(0,0,0,0.35) 0%, rgba(0,0,0,0.05) 100%);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 0.875rem;
}

.ad-text {
  display: flex;
  flex-direction: column;
}

.ad-label {
  font-size: 0.625rem;
  color: rgba(255, 255, 255, 0.8);
  margin-bottom: 0.0625rem;
}

.ad-title {
  font-size: 0.9375rem;
  font-weight: 700;
  color: #fff;
  margin-bottom: 0.0625rem;
}

.ad-desc {
  font-size: 0.6875rem;
  color: rgba(255, 255, 255, 0.85);
}

.ad-btn {
  padding: 0.3125rem 0.75rem;
  border-radius: 1rem;
  flex-shrink: 0;

  text {
    color: #fff;
    font-size: 0.75rem;
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
  gap: 0.75rem;
}

.lost-found-card {
  width: 9.375rem;
  flex-shrink: 0;
  background-color: $uni-bg-color-card;
  border-radius: 0.4375rem;
  overflow: hidden;
  box-shadow: $uni-shadow-sm;
}

.lost-found-image-wrap {
  width: 100%;
  height: 6.25rem;
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
  font-size: 2.5rem;
  opacity: 0.4;
}

.lost-found-tag {
  position: absolute;
  top: 0.375rem;
  left: 0.375rem;
  border-radius: 0.25rem;
  padding: 0.125rem 0.4375rem;
}

.lost-found-tag-text {
  font-size: 0.625rem;
  color: #fff;
}

.lost-found-body {
  padding: 0.4375rem 0.5rem 0.5625rem;
}

.lost-found-title {
  display: block;
  font-size: 0.8125rem;
  font-weight: 500;
  color: $uni-text-color;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  margin-bottom: 0.125rem;
}

.lost-found-time {
  font-size: 0.625rem;
  color: $uni-text-color-grey;
}
</style>
