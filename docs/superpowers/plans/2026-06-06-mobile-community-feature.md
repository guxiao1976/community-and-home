# Mobile Community Management Feature Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add community membership management (Pinia store), visual community switcher on notice page, redesign three homepage sections (notices, contacts, lost&found), and integrate with join-community and mine pages.

**Architecture:** A new `communityStore` Pinia store manages user's community memberships and current active community (persisted in localStorage). The notice page has a header greeting + community switcher that triggers data reload. All API calls pass `communityId` from the store instead of a hardcoded constant.

**Tech Stack:** Vue 3 + TypeScript + Uni-app + Pinia + Axios (lossless-json), SCSS with warm coffee color scheme

---

### Task 1: Create Community Pinia Store

**Files:**
- Create: `src/stores/community.ts`

- [ ] **Step 1: Write the store file**

```typescript
// src/stores/community.ts
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { getUserMemberships } from '@/api/user';
import type { CommunityMembership } from '@/api/user';

export interface CommunityInfo {
  communityId: string;
  communityName: string;
  address?: string;
}

function loadStoredCommunityId(): string {
  try {
    return uni.getStorageSync('current_community_id') || '';
  } catch {
    return '';
  }
}

function saveStoredCommunityId(id: string): void {
  try {
    uni.setStorageSync('current_community_id', id);
  } catch {
    // ignore storage errors
  }
}

export const useCommunityStore = defineStore('community', () => {
  // --- State ---
  const communities = ref<CommunityInfo[]>([]);
  const currentCommunityId = ref<string>(loadStoredCommunityId());

  // --- Getters ---
  const currentCommunity = computed<CommunityInfo | null>(() => {
    if (!currentCommunityId.value) return null;
    return communities.value.find(c => c.communityId === currentCommunityId.value) || null;
  });

  const hasCommunities = computed(() => communities.value.length > 0);

  const communityCount = computed(() => communities.value.length);

  // --- Actions ---
  async function loadMemberships(): Promise<void> {
    try {
      const memberships = await getUserMemberships();
      communities.value = memberships.map((m: CommunityMembership) => ({
        communityId: m.communityId,
        communityName: (m as any).communityName || ('小区 ' + m.communityId),
        address: (m as any).address || undefined,
      }));

      // If no current selection (first load), pick first community
      if (!currentCommunityId.value && communities.value.length > 0) {
        currentCommunityId.value = communities.value[0].communityId;
        saveStoredCommunityId(currentCommunityId.value);
      }
    } catch {
      communities.value = [];
    }
  }

  function switchCommunity(id: string): void {
    const exists = communities.value.some(c => c.communityId === id);
    if (!exists) return;
    currentCommunityId.value = id;
    saveStoredCommunityId(id);
  }

  function addCommunity(membership: { communityId: string; communityName?: string; address?: string }): void {
    const exists = communities.value.some(c => c.communityId === membership.communityId);
    if (exists) return;

    communities.value.push({
      communityId: membership.communityId,
      communityName: membership.communityName || ('小区 ' + membership.communityId),
      address: membership.address,
    });

    // Auto-select newly joined community
    currentCommunityId.value = membership.communityId;
    saveStoredCommunityId(membership.communityId);
  }

  return {
    // State
    communities,
    currentCommunityId,
    // Getters
    currentCommunity,
    hasCommunities,
    communityCount,
    // Actions
    loadMemberships,
    switchCommunity,
    addCommunity,
  };
});
```

---

### Task 2: Update Community API — Remove Hardcoded COMMUNITY_ID

**Files:**
- Modify: `src/api/community.ts`

- [ ] **Step 1: Remove the hardcoded COMMUNITY_ID constant and make communityId a required parameter**

Remove the line:
```typescript
const COMMUNITY_ID = '1'; // TODO: fetch from user-service default community
```

And update function signatures to remove default values:

- `getNoticeList(communityId: string, page: number = 1, pageSize: number = 3)` — remove `= COMMUNITY_ID` from communityId
- `getContacts(communityId: string)` — remove `= COMMUNITY_ID` from communityId
- `getLostFoundList(communityId: string, page: number = 1, pageSize: number = 3)` — remove `= COMMUNITY_ID` from communityId

---

### Task 3: Rewrite notice.vue — Community Switcher + Visual Upgrade

**Files:**
- Modify: `src/pages/notice/notice.vue`

- [ ] **Step 1: Rewrite notice.vue completely**

Key changes:
- Add greeting header section with community switcher (before the three sections)
- Community switcher shows current community name with dropdown (uni.showActionSheet)
- If no communities, show "请先加入小区" guidance linking to `/pages/join-community/join-community`
- Greeting: 早安/午安/晚安 based on current hour, using nickname from useUserStore
- Notice cards: left colored vertical bar + pill badge for role tag
- Contacts: 2-column grid layout with phone click
- Lost&found: image card horizontal scroll with type tag overlay
- Use `useCommunityStore` for communityId, pass to all API calls
- watch(currentCommunityId) to reload data on switch
- onMounted: load memberships then load three sections

Full implementation code for notice.vue (template + script + style):

```vue
<template>
  <view class="page">
    <!-- Community Switcher Header -->
    <view class="header">
      <text class="greeting">{{ greetingText }}{{ userStore.nickname ? '，' + userStore.nickname : '' }}</text>

      <view v-if="communityStore.hasCommunities" class="switcher" @click="onSwitchCommunity">
        <text class="switcher-icon">📍</text>
        <text class="switcher-name">{{ communityStore.currentCommunity?.communityName || '选择小区' }}</text>
        <text class="switcher-arrow">▼</text>
      </view>
      <view v-else class="no-community" @click="goJoinCommunity">
        <text class="no-community-icon">📍</text>
        <text class="no-community-text">请先加入小区</text>
        <text class="no-community-arrow">→</text>
      </view>

      <text class="page-title">公告信息</text>
    </view>

    <!-- Section A: 通知公告 -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">📢 通知公告</text>
        <text class="section-more" @click="onMoreNotice">更多 ▸</text>
      </view>
      <view v-if="notices.length === 0" class="empty-state">
        <text class="empty-text">暂无通知</text>
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
              <view class="notice-role-pill" :style="{ backgroundColor: getNoticeRoleColor(item.role) + '20', borderColor: getNoticeRoleColor(item.role) }">
                <text class="role-text" :style="{ color: getNoticeRoleColor(item.role) }">{{ getNoticeRoleName(item.role) }}</text>
              </view>
            </view>
            <text class="notice-time">{{ formatTime(item.publishedAt || item.createdAt) }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Section B: 便民联络 -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">📞 便民联络</text>
      </view>
      <view v-if="contactGroups.length === 0" class="empty-state">
        <text class="empty-text">暂无联络信息</text>
      </view>
      <view v-else class="contact-grid">
        <view
          v-for="(group, idx) in contactGroups"
          :key="idx"
          class="contact-card"
        >
          <text class="contact-category-icon">{{ group.icon }}</text>
          <text class="contact-category-name">{{ group.categoryName }}</text>
          <view
            v-for="contact in group.items"
            :key="contact.id"
            class="contact-phone-row"
            @click="onCallPhone(contact.phone)"
          >
            <text class="contact-phone">{{ contact.phone }}</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Section C: 寻失互助 -->
    <view class="section">
      <view class="section-header">
        <text class="section-title">🔍 寻失互助</text>
      </view>
      <view v-if="lostFoundItems.length === 0" class="empty-state">
        <text class="empty-text">暂无寻失信息</text>
      </view>
      <scroll-view v-else scroll-x class="lost-found-scroll">
        <view class="lost-found-list">
          <view
            v-for="item in lostFoundItems"
            :key="item.id"
            class="lost-found-card"
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
              <view class="lost-found-tag-overlay" :style="{ backgroundColor: item.type === 1 ? '#B8956A' : '#8DAF7E' }">
                <text class="tag-overlay-text">{{ getLostFoundTypeName(item.type) }}</text>
              </view>
            </view>
            <text class="lost-found-title">{{ item.title }}</text>
          </view>
        </view>
      </scroll-view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
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
import dayjs from 'dayjs';

const userStore = useUserStore();
const communityStore = useCommunityStore();

// --- Greeting ---
const greetingText = computed(() => {
  const hour = new Date().getHours();
  if (hour < 12) return '早安';
  if (hour < 18) return '午安';
  return '晚安';
});

// --- Notice State ---
const notices = ref<Notice[]>([]);

// --- Contact State ---
const contacts = ref<Contact[]>([]);

interface ContactGroup {
  category: number;
  categoryName: string;
  icon: string;
  items: Contact[];
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
    groups.push({
      category: cat,
      categoryName: getContactCategoryName(cat),
      icon: getContactCategoryIcon(cat),
      items,
    });
  }
  return groups;
});

// --- Lost & Found State ---
const lostFoundItems = ref<LostFoundItem[]>([]);
const imageErrors = ref<Set<string>>(new Set());

// --- Actions ---
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

function onSwitchCommunity() {
  const names = communityStore.communities.map(c => c.communityName);
  uni.showActionSheet({
    itemList: names,
    success: (res) => {
      const selected = communityStore.communities[res.tapIndex];
      if (selected) {
        communityStore.switchCommunity(selected.communityId);
      }
    },
  });
}

function goJoinCommunity() {
  uni.navigateTo({ url: '/pages/join-community/join-community' });
}

function formatTime(ts: number): string {
  if (!ts) return '';
  return dayjs.unix(ts).format('YYYY-MM-DD HH:mm');
}

// --- Fetch Data ---
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
  fetchNotices();
  fetchContacts();
  fetchLostFound();
}

// --- Lifecycle ---
watch(() => communityStore.currentCommunityId, (newVal, oldVal) => {
  if (newVal && newVal !== oldVal) {
    loadAll();
  }
});

onMounted(async () => {
  await communityStore.loadMemberships();
  loadAll();
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
  padding: 0 32rpx;
  padding-bottom: calc(100rpx + env(safe-area-inset-bottom));
}

// ---- Header ----
.header {
  padding: 40rpx 0 24rpx;
}

.greeting {
  display: block;
  font-size: 34rpx;
  font-weight: 700;
  color: $uni-text-color;
  margin-bottom: 20rpx;
}

.switcher {
  display: inline-flex;
  align-items: center;
  background-color: $uni-bg-color-card;
  border-radius: 16rpx;
  padding: 16rpx 24rpx;
  box-shadow: $uni-shadow-base;
  margin-bottom: 24rpx;
}

.switcher-icon {
  font-size: 28rpx;
  margin-right: 8rpx;
}

.switcher-name {
  font-size: 28rpx;
  font-weight: 500;
  color: $uni-text-color;
  margin-right: 8rpx;
}

.switcher-arrow {
  font-size: 20rpx;
  color: $uni-text-color-grey;
}

.no-community {
  display: inline-flex;
  align-items: center;
  background-color: $uni-bg-color-card;
  border-radius: 16rpx;
  padding: 16rpx 24rpx;
  box-shadow: $uni-shadow-base;
  margin-bottom: 24rpx;
}

.no-community-icon {
  font-size: 28rpx;
  margin-right: 8rpx;
}

.no-community-text {
  font-size: 28rpx;
  color: $uni-color-primary;
  margin-right: 8rpx;
}

.no-community-arrow {
  font-size: 24rpx;
  color: $uni-color-primary;
}

.page-title {
  display: block;
  font-size: 34rpx;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 0;
}

// ---- Sections ----
.section {
  margin-bottom: 32rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.section-title {
  font-size: 32rpx;
  font-weight: 600;
  color: $uni-text-color;
}

.section-more {
  font-size: 24rpx;
  color: $uni-color-primary;
}

.empty-state {
  display: flex;
  justify-content: center;
  padding: 48rpx 0;
}

.empty-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}

// ---- Notice Cards (left colored bar + pill badge) ----
.notice-list {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
}

.notice-card {
  display: flex;
  background-color: $uni-bg-color-card;
  border-radius: 16rpx;
  overflow: hidden;
  box-shadow: $uni-shadow-base;
}

.notice-bar {
  width: 8rpx;
  flex-shrink: 0;
}

.notice-body {
  flex: 1;
  padding: 24rpx;
}

.notice-card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10rpx;
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
  border-radius: 24rpx;
  padding: 4rpx 14rpx;
  border: 1rpx solid;
  flex-shrink: 0;
}

.role-text {
  font-size: 20rpx;
}

.notice-time {
  font-size: 24rpx;
  color: $uni-text-color-grey;
}

// ---- Contact Grid ----
.contact-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx;
}

.contact-card {
  background-color: $uni-bg-color-card;
  border-radius: 16rpx;
  padding: 24rpx;
  box-shadow: $uni-shadow-base;
}

.contact-category-icon {
  font-size: 36rpx;
  display: block;
  margin-bottom: 8rpx;
}

.contact-category-name {
  font-size: 26rpx;
  font-weight: 500;
  color: $uni-text-color;
  display: block;
  margin-bottom: 12rpx;
}

.contact-phone-row {
  padding: 8rpx 0;
  border-top: 1rpx solid $uni-border-color;

  &:first-of-type {
    border-top: none;
  }
}

.contact-phone {
  font-size: 24rpx;
  color: $uni-color-primary;
}

// ---- Lost & Found ----
.lost-found-scroll {
  white-space: nowrap;
}

.lost-found-list {
  display: inline-flex;
  gap: 20rpx;
}

.lost-found-card {
  width: 240rpx;
  flex-shrink: 0;
  background-color: $uni-bg-color-card;
  border-radius: 16rpx;
  overflow: hidden;
  box-shadow: $uni-shadow-base;
}

.lost-found-image-wrap {
  width: 100%;
  height: 180rpx;
  background-color: $uni-bg-color-grey;
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
  font-size: 56rpx;
}

.lost-found-tag-overlay {
  position: absolute;
  top: 12rpx;
  left: 12rpx;
  border-radius: 8rpx;
  padding: 4rpx 12rpx;
}

.tag-overlay-text {
  font-size: 20rpx;
  color: #FFFFFF;
}

.lost-found-title {
  display: block;
  padding: 12rpx 12rpx 16rpx;
  font-size: 26rpx;
  color: $uni-text-color;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
```

---

### Task 4: Update join-community.vue — Update Store on Join

**Files:**
- Modify: `src/pages/join-community/join-community.vue`

- [ ] **Step 1: Import communityStore and call addCommunity after successful join**

Add import at top of `<script setup>`:
```typescript
import { useCommunityStore } from '@/stores/community';
```

Inside `joinCommunity` function, after successful join (`const mem = await joinCommunityApi(area.id);`), add:
```typescript
const communityStore = useCommunityStore();
communityStore.addCommunity({
  communityId: area.id,
  communityName: area.name,
  address: area.address,
});
```

---

### Task 5: Update mine.vue — Show Community Info

**Files:**
- Modify: `src/pages/mine/mine.vue`

- [ ] **Step 1: Add community membership count display in the logged-in profile section**

Add after the nickname line in the template (inside the `v-else` block):
```vue
<view class="community-info" v-if="communityStore.communityCount > 0">
  <text class="community-count">已加入 {{ communityStore.communityCount }}/3 个小区</text>
</view>
<view class="community-info" v-else>
  <text class="community-count">暂未加入小区</text>
</view>
```

Add script import:
```typescript
import { useCommunityStore } from '@/stores/community';
const communityStore = useCommunityStore();
```

Add onMounted to load memberships:
```typescript
import { onMounted } from 'vue';
onMounted(() => {
  if (userStore.isLoggedIn) {
    communityStore.loadMemberships();
  }
});
```

Add styles:
```scss
.community-info {
  margin-top: 12rpx;
}

.community-count {
  font-size: 26rpx;
  color: $uni-color-primary;
}
```

---

### Task 6: Build Verification

- [ ] **Step 1: Run `npm run build:h5`**

```bash
cd /home/jiaoxh/my-project/community-home/web/mobile && npm run build:h5
```

Expected: Build succeeds with no TypeScript errors.

---
