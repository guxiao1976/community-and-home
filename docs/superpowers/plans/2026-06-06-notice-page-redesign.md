# 公告信息页面重设计 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将移动端公告信息页面从简陋的基础样式重设计为温暖社区风格的精致页面，并实现自定义小区下拉切换器

**Architecture:** 新建 `CommunitySwitcher` 组件封装下拉面板逻辑，重写 `notice.vue` 页面（template/style/script），复用现有 `communityStore` 和 `api/community.ts`

**Tech Stack:** Uni-app (Vue 3) + TypeScript + Pinia + SCSS

**Spec:** `docs/superpowers/specs/2026-06-06-notice-page-redesign.md`

---

## 文件结构

```
web/mobile/src/
├── pages/notice/
│   └── notice.vue                    ← 重写（template + style + script）
└── components/
    └── community-switcher.vue         ← 新建（小区下拉切换器组件）
```

- `community-switcher.vue`：封装头部区域（渐变背景 + 小区选择器 + 下拉面板），通过 emit 向父组件抛 `switch` 事件
- `notice.vue`：三个内容板块 + 广告位 + 骨架屏 + 空状态，通过 `communityStore` 管理小区切换与数据刷新

---

## 实现任务

### Task 1: 创建 CommunitySwitcher 组件

**Files:**
- Create: `web/mobile/src/components/community-switcher.vue`

- [ ] **Step 1: 创建组件文件，实现 template**

```vue
<template>
  <view class="cs-root">
    <!-- 渐变头部背景 -->
    <view class="cs-header">
      <!-- 切换器按钮 -->
      <view class="cs-trigger" @click.stop="toggle">
        <text class="cs-trigger-icon">🏘️</text>
        <view class="cs-trigger-body">
          <text class="cs-trigger-label">当前小区</text>
          <text class="cs-trigger-name">{{ currentName }}</text>
        </view>
        <view v-if="communityCount > 0" class="cs-badge">
          <text>{{ communityCount }}个小区</text>
        </view>
        <text class="cs-arrow" :class="{ 'cs-arrow--up': open }">▼</text>
      </view>
    </view>

    <!-- 遮罩层 -->
    <view v-if="open" class="cs-overlay" @click="close" />

    <!-- 下拉面板 -->
    <view v-if="open" class="cs-dropdown">
      <view class="cs-dropdown-header">
        <text class="cs-dropdown-header-icon">🏘️</text>
        <text class="cs-dropdown-header-text">已加入的小区 ({{ communityCount }}个)</text>
      </view>
      <scroll-view class="cs-dropdown-list" scroll-y>
        <view
          v-for="c in communities"
          :key="c.communityId"
          class="cs-dropdown-item"
          :class="{ 'cs-dropdown-item--active': c.communityId === modelValue }"
          @click="select(c.communityId)"
        >
          <text class="cs-dropdown-item-icon">🏘️</text>
          <view class="cs-dropdown-item-body">
            <text class="cs-dropdown-item-name">{{ c.communityName }}</text>
            <text v-if="c.address" class="cs-dropdown-item-addr">{{ c.address }}</text>
          </view>
          <text v-if="c.communityId === modelValue" class="cs-dropdown-item-check">✓</text>
        </view>
      </scroll-view>
    </view>
  </view>
</template>
```

- [ ] **Step 2: 实现 script 逻辑**

```typescript
<script setup lang="ts">
import { ref, computed } from 'vue';
import type { CommunityInfo } from '@/stores/community';

const props = defineProps<{
  communities: CommunityInfo[];
  modelValue: string;
}>();

const emit = defineEmits<{
  'update:modelValue': [id: string];
  switch: [id: string];
}>();

const open = ref(false);

const communityCount = computed(() => props.communities.length);

const currentName = computed(() => {
  const c = props.communities.find(x => x.communityId === props.modelValue);
  return c?.communityName || '选择小区';
});

function toggle() {
  if (props.communities.length === 0) return;
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function select(id: string) {
  emit('update:modelValue', id);
  emit('switch', id);
  open.value = false;
}
</script>
```

- [ ] **Step 3: 实现 style（SCSS）**

```scss
<style scoped lang="scss">
.cs-root {
  position: relative;
  z-index: 100;
}

.cs-header {
  background: linear-gradient(160deg, $uni-color-primary-light 0%, #E8DCCF 25%, $uni-bg-color-card 50%, $uni-bg-color 75%);
  padding: 24rpx 32rpx 18rpx;
}

.cs-trigger {
  display: flex;
  align-items: center;
  background: rgba(255, 255, 255, 0.88);
  border-radius: 12rpx;
  padding: 16rpx 20rpx;
  border: 1px solid $uni-border-color;
  box-shadow: $uni-shadow-sm;
  backdrop-filter: blur(10px);
}

.cs-trigger-icon {
  font-size: 36rpx;
  margin-right: 14rpx;
  flex-shrink: 0;
}

.cs-trigger-body {
  flex: 1;
  min-width: 0;
}

.cs-trigger-label {
  display: block;
  font-size: 20rpx;
  color: $uni-text-color-grey;
  margin-bottom: 2rpx;
}

.cs-trigger-name {
  display: block;
  font-size: 30rpx;
  font-weight: 600;
  color: $uni-text-color;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.cs-badge {
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
  background: rgba($uni-color-primary, 0.1);
  margin-right: 10rpx;
  flex-shrink: 0;

  text {
    font-size: 22rpx;
    color: $uni-color-primary;
    font-weight: 500;
  }
}

.cs-arrow {
  font-size: 20rpx;
  color: $uni-text-color-grey;
  transition: transform 0.2s;
  flex-shrink: 0;

  &--up {
    transform: rotate(180deg);
  }
}

.cs-overlay {
  position: fixed;
  inset: 0;
  z-index: 199;
  background: rgba(0, 0, 0, 0.25);
}

.cs-dropdown {
  position: absolute;
  top: calc(100% - 4rpx);
  left: 32rpx;
  right: 32rpx;
  z-index: 200;
  background: #fff;
  border-radius: 14rpx;
  box-shadow: 0 8rpx 40rpx rgba(0, 0, 0, 0.1);
  border: 1px solid $uni-border-color;
  overflow: hidden;
}

.cs-dropdown-header {
  display: flex;
  align-items: center;
  gap: 8rpx;
  padding: 20rpx 28rpx;
  border-bottom: 1px solid $uni-bg-color-grey;
}

.cs-dropdown-header-icon {
  font-size: 24rpx;
}

.cs-dropdown-header-text {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

.cs-dropdown-list {
  max-height: 400rpx;
}

.cs-dropdown-item {
  display: flex;
  align-items: center;
  padding: 20rpx 28rpx;

  &--active {
    background: rgba($uni-color-primary, 0.05);

    .cs-dropdown-item-name {
      color: $uni-color-primary;
      font-weight: 600;
    }
  }
}

.cs-dropdown-item-icon {
  font-size: 36rpx;
  margin-right: 16rpx;
  flex-shrink: 0;
}

.cs-dropdown-item-body {
  flex: 1;
  min-width: 0;
}

.cs-dropdown-item-name {
  display: block;
  font-size: 28rpx;
  color: $uni-text-color;
}

.cs-dropdown-item-addr {
  display: block;
  font-size: 22rpx;
  color: $uni-text-color-grey;
  margin-top: 2rpx;
}

.cs-dropdown-item-check {
  font-size: 34rpx;
  color: $uni-color-primary;
  font-weight: 700;
  flex-shrink: 0;
  margin-left: 16rpx;
}
</style>
```

- [ ] **Step 4: 验证组件编译**

```bash
cd /home/jiaoxh/my-project/community-home/web/mobile && npx vue-tsc --noEmit --project tsconfig.app.json 2>&1 | head -30
```

---

### Task 2: 重写 notice.vue — 页面模板

**Files:**
- Modify: `web/mobile/src/pages/notice/notice.vue`

**说明：** 完全重写 template，引入 CommunitySwitcher 组件，重新组织三个板块的 HTML 结构，加入骨架屏和空状态。

- [ ] **Step 5: 重写 template**

```vue
<template>
  <view class="page">
    <!-- 小区切换器（含渐变头部） -->
    <CommunitySwitcher
      :communities="communityStore.communities"
      :model-value="communityStore.currentCommunityId"
      @switch="onCommunitySwitch"
    />

    <!-- 骨架屏 -->
    <template v-if="loading">
      <view class="skeleton-wrap">
        <!-- 通知骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title" />
          <view class="skeleton-notice" v-for="i in 3" :key="'sn'+i">
            <view class="skeleton-bar" />
            <view class="skeleton-text" />
          </view>
        </view>
        <!-- 联络骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title short" />
          <view class="skeleton-grid">
            <view class="skeleton-contact" v-for="i in 4" :key="'sc'+i" />
          </view>
        </view>
        <!-- 寻失骨架 -->
        <view class="skeleton-section">
          <view class="skeleton-title short" />
          <view class="skeleton-lf-row">
            <view class="skeleton-lf-card" v-for="i in 2" :key="'slf'+i" />
          </view>
        </view>
      </view>
    </template>

    <!-- 实际内容 -->
    <template v-else>
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
      <view class="ad-banner" style="background-image: url('https://images.unsplash.com/photo-1542838132-92c53300491e?w=800&h=200&fit=crop');">
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
      <view class="ad-banner" style="background-image: url('https://images.unsplash.com/photo-1581578731548-c64695cc6952?w=800&h=200&fit=crop');">
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
      <view class="ad-banner" style="background-image: url('https://images.unsplash.com/photo-1523050854058-8df90910b48a?w=800&h=200&fit=crop');">
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
```

---

### Task 3: 重写 notice.vue — 脚本逻辑

**Files:**
- Modify: `web/mobile/src/pages/notice/notice.vue`（script 部分）

- [ ] **Step 6: 重写 script setup**

```typescript
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
  uni.showToast({ title: '广告详情开发中...', icon: 'none', duration: 2000 });
}

function onCommunitySwitch(id: string) {
  communityStore.switchCommunity(id);
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
```

---

### Task 4: 重写 notice.vue — 样式

**Files:**
- Modify: `web/mobile/src/pages/notice/notice.vue`（style 部分）

- [ ] **Step 7: 重写 style（SCSS）**

```scss
<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: $uni-bg-color;
  padding-bottom: calc(100rpx + env(safe-area-inset-bottom));
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

  & + .ad-banner {
    margin-top: 24rpx;
  }

  &:last-of-type {
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
```

---

### Task 5: 类型检查与验证

**Files:**
- 无新建文件

- [ ] **Step 8: 运行 TypeScript 类型检查**

```bash
cd /home/jiaoxh/my-project/community-home/web/mobile && npx vue-tsc --noEmit --project tsconfig.app.json 2>&1 | head -50
```

**预期：** 无新增类型错误（可能有已存在的第三方库类型警告）

- [ ] **Step 9: 启动 dev server 验证编译**

```bash
cd /home/jiaoxh/my-project/community-home/web/mobile && npm run dev:h5
```

**验证：**
1. 浏览器打开 H5 页面
2. 确认头部渐变 + 小区切换器显示正常
3. 点击切换器展开下拉面板，确认小区列表显示正确
4. 确认通知、联络、寻失三个板块渲染正常
5. 确认广告位图片加载显示
6. 确认点击电话号码可拨号

- [ ] **Step 10: Commit**

```bash
cd /home/jiaoxh/my-project/community-home
git add web/mobile/src/components/community-switcher.vue
git add web/mobile/src/pages/notice/notice.vue
git commit -m "feat(mobile): redesign notice page with community dropdown switcher

- Replace basic layout with warm community visual style
- Add CommunitySwitcher component with custom dropdown panel
- Redesign notices, contacts (2-col grid), lost & found sections
- Add 3 ad banner slots (2 after contacts, 1 after lost & found)
- Add skeleton screen loading animation
- Add friendly empty state illustrations
- Support pull-to-refresh

Closes: notice page redesign"
```

---

## 验证清单

| # | 验证项 | 方法 |
|---|--------|------|
| 1 | 小区切换器下拉面板正常展开/收起 | 点击切换器，点击遮罩关闭 |
| 2 | 切换小区后数据刷新 | 选择不同小区，确认列表更新 |
| 3 | 骨架屏 → 内容过渡 | 刷新页面，观察 loading 状态 |
| 4 | 空状态显示 | 无数据小区确认空状态展示 |
| 5 | 电话号码点击拨号 | 点击联络卡片，确认跳转拨号 |
| 6 | 广告位渲染 | 确认3个广告横幅正常显示 |
| 7 | 下拉刷新 | 页面下拉，确认数据重新加载 |
| 8 | TypeScript 无类型错误 | `npx vue-tsc --noEmit` |
