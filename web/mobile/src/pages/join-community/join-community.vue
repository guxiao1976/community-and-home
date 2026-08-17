<template>
  <view class="page">
    <!-- Header -->
    <view class="header">
      <text class="header-title">加入小区</text>
      <text class="header-sub">选择您的小区，开始社区生活</text>
    </view>

    <!-- Max Limit Banner -->
    <view v-if="maxReached" class="max-banner">
      <text>⚠️ 您已加入 {{ activeCount }}/3 个小区，已达上限。如需加入新小区，请先退出已有小区。</text>
    </view>

    <!-- My Communities (always visible, loaded from store) -->
    <view v-if="storeCommunities.length > 0" class="my-section-top">
      <text class="my-title">已加入 {{ activeCount }}/3 个小区</text>
      <view class="my-tags">
        <view v-for="c in storeCommunities" :key="c.communityId" class="my-badge">
          <text>{{ c.communityName }}</text>
        </view>
      </view>
    </view>

    <!-- Step Indicator -->
    <view class="steps">
      <view class="step" :class="{ active: step === 1, done: step > 1 }">
        <view class="step-num">{{ step > 1 ? '✓' : '1' }}</view>
        <text class="step-label">选择省份</text>
      </view>
      <view class="step-line" :class="{ done: step > 1 }" />
      <view class="step" :class="{ active: step === 2, done: step > 2 }">
        <view class="step-num">{{ step > 2 ? '✓' : '2' }}</view>
        <text class="step-label">选择城市</text>
      </view>
      <view class="step-line" :class="{ done: step > 2 }" />
      <view class="step" :class="{ active: step === 3, done: step > 3 }">
        <view class="step-num">{{ step > 3 ? '✓' : '3' }}</view>
        <text class="step-label">选择县区</text>
      </view>
      <view class="step-line" :class="{ done: step > 3 }" />
      <view class="step" :class="{ active: step === 4 }">
        <view class="step-num">4</view>
        <text class="step-label">搜索小区</text>
      </view>
    </view>

    <!-- Step 1: Select Province -->
    <view v-if="step === 1" class="card">
      <text class="card-title">请选择省份</text>
      <view v-if="maxReached" class="max-inline">
        <text>已达加入上限（3个），无法选择。</text>
      </view>
      <view v-else-if="provincesLoading" class="loading-text">加载中...</view>
      <scroll-view v-else class="list" scroll-y>
        <view
          v-for="p in provinces"
          :key="p.id"
          class="list-item"
          :class="{ selected: selectedProvince?.id === p.id }"
          @click="selectProvince(p)"
        >
          <text class="item-name">{{ p.name }}</text>
          <text v-if="selectedProvince?.id === p.id" class="item-check">✓</text>
        </view>
      </scroll-view>
    </view>

    <!-- Step 2: Select City -->
    <view v-if="step === 2" class="card">
      <view class="back-row" @click="step = 1">
        <text class="back-icon">←</text>
        <text class="back-text">{{ selectedProvince?.name }}</text>
      </view>
      <text class="card-title">请选择城市</text>
      <view v-if="citiesLoading" class="loading-text">加载中...</view>
      <scroll-view v-else class="list" scroll-y>
        <view
          v-for="c in cities"
          :key="c.id"
          class="list-item"
          :class="{ selected: selectedCity?.id === c.id }"
          @click="selectCity(c)"
        >
          <text class="item-name">{{ c.name }}</text>
          <text v-if="selectedCity?.id === c.id" class="item-check">✓</text>
        </view>
      </scroll-view>
    </view>

    <!-- Step 3: Select District -->
    <view v-if="step === 3" class="card">
      <view class="back-row" @click="step = 2">
        <text class="back-icon">←</text>
        <text class="back-text">{{ selectedCity?.name }}</text>
      </view>
      <text class="card-title">请选择县区</text>
      <view v-if="districtsLoading" class="loading-text">加载中...</view>
      <scroll-view v-else class="list" scroll-y>
        <view
          v-for="d in districts"
          :key="d.id"
          class="list-item"
          :class="{ selected: selectedDistrict?.id === d.id }"
          @click="selectDistrict(d)"
        >
          <text class="item-name">{{ d.name }}</text>
          <text v-if="selectedDistrict?.id === d.id" class="item-check">✓</text>
        </view>
      </scroll-view>
      <button class="btn" :class="{ 'btn--disabled': !selectedDistrict }" :disabled="!selectedDistrict" @click="goSearch">
        搜索小区
      </button>
    </view>

    <!-- Step 4: Search & Join -->
    <view v-if="step === 4" class="card">
      <view class="back-row" @click="step = 3">
        <text class="back-icon">←</text>
        <text class="back-text">{{ selectedProvince?.name }} · {{ selectedCity?.name }} · {{ selectedDistrict?.name }}</text>
      </view>

      <!-- Search -->
      <view class="search-row">
        <input v-model="keyword" class="search-input" placeholder="输入小区名称模糊搜索" @confirm="doSearch" />
        <view class="search-btn" @click="doSearch"><text>搜索</text></view>
      </view>

      <!-- Results -->
      <view v-if="searching" class="loading-text">搜索中...</view>
      <scroll-view v-else-if="areas.length > 0" class="list" scroll-y>
        <view v-for="area in areas" :key="area.id" class="list-item" @click="onSelectArea(area)">
          <view class="item-content">
            <text class="item-name">{{ area.name }}</text>
            <text v-if="area.address" class="item-addr">{{ area.address }}</text>
          </view>
          <view class="join-tag" v-if="!isJoined(area.id)">加入</view>
          <text v-else class="joined-text">已加入</text>
        </view>
      </scroll-view>
      <view v-else-if="!searching && searched" class="empty-text">未找到匹配的小区</view>
      <view v-else-if="!searching && !searched" class="loading-text">输入名称后点击搜索</view>
    </view>

    <!-- Max Warning -->
    <view v-if="showMaxWarning" class="max-warning"><text>最多加入 3 个小区</text></view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import {
  type Division, type ResidentialArea,
  getDivisions, searchResidentialAreas, joinCommunity,
} from '@/api/user';
import { useCommunityStore } from '@/stores/community';
// SEE: [[frontend-cross-page-storage-contract]] — 选中小区存 pending-join 唯一契约源，join-choice/join-residence 消费
import { savePendingJoin } from '@/utils/pending-join';

const communityStore = useCommunityStore();

// Active count and max check (from store, only bind_status=1)
const activeCount = computed(() => communityStore.communityCount);
const maxReached = computed(() => activeCount.value >= 3);
// Store communities have preserved names from previous loads
const storeCommunities = computed(() => communityStore.communities);

const step = ref(1);

// Step 1: Provinces (level 1)
const provinces = ref<Division[]>([]);
const selectedProvince = ref<Division | null>(null);
const provincesLoading = ref(false);

// Step 2: Cities (level 2)
const cities = ref<Division[]>([]);
const selectedCity = ref<Division | null>(null);
const citiesLoading = ref(false);

// Step 3: Districts (level 3)
const districts = ref<Division[]>([]);
const selectedDistrict = ref<Division | null>(null);
const districtsLoading = ref(false);

// Step 4: Search
const keyword = ref('');
const areas = ref<ResidentialArea[]>([]);
const searching = ref(false);
const searched = ref(false);

// Memberships (from store, loaded on mount)
const showMaxWarning = ref(false);


onMounted(async () => {
  provincesLoading.value = true;
  try {
    // Load store first to get community names, then divisions
    await Promise.all([
      communityStore.loadMemberships().catch(() => {}),
      getDivisions().then(d => { provinces.value = d; }).catch(() => {}),
    ]);
  } catch (_) { /* ignore */ }
  finally { provincesLoading.value = false; }
});

async function selectProvince(p: Division) {
  selectedProvince.value = p;
  selectedCity.value = null;
  selectedDistrict.value = null;
  citiesLoading.value = true;
  try { cities.value = await getDivisions(p.id); } catch (_) { cities.value = []; }
  finally { citiesLoading.value = false; }
  // 自动级联：选完省立即进入城市列表（无需点下一步）
  step.value = 2;
}

async function selectCity(c: Division) {
  selectedCity.value = c;
  selectedDistrict.value = null;
  districtsLoading.value = true;
  try { districts.value = await getDivisions(c.id); } catch (_) { districts.value = []; }
  finally { districtsLoading.value = false; }
  // 自动级联：选完市立即进入县区列表（无需点下一步）
  step.value = 3;
}

function selectDistrict(d: Division) {
  selectedDistrict.value = d;
}

function goSearch() {
  step.value = 4;
  if (keyword.value) doSearch();
}

async function doSearch() {
  if (!keyword.value.trim() || !selectedDistrict.value) return;
  searching.value = true;
  searched.value = true;
  try {
    areas.value = await searchResidentialAreas({
      keyword: keyword.value.trim(),
      countyId: selectedDistrict.value.id,
    });
  } catch (_) { areas.value = []; }
  finally { searching.value = false; }
}

// --- 加入流程重构（用户拍板）：点「加入小区」= 立即建 membership（无房号）---
// 选中小区 → await joinCommunity(communityId)（房号可选，后端并行管线已支持）→
// communityStore.addCommunity 立即进 store → 把 membership.id 回填 pending-join（新增 membershipId）
// → 导航 join-choice（已加入 XX，请选择下一步：填房号成为业主 / 其他身份认证）。
// 已加入 → toast 不重复；maxReached → 上限警告；joinCommunity 失败 → toast 错误（可重试）。
async function onSelectArea(area: ResidentialArea) {
  if (maxReached.value) {
    showMaxWarning.value = true;
    setTimeout(() => { showMaxWarning.value = false; }, 2000);
    return;
  }
  if (isJoined(area.id)) {
    uni.showToast({ title: '该小区已加入', icon: 'none' });
    return;
  }
  // 前端仅做前置 UX 拦截 + 错误透出；加入上限/重复等由后端权威校验（拦截器已 toast，此处按 msg 透出兜底）
  // SEE: [[frontend-business-rule-hardcode]] [[testing-discipline]]
  try {
    const membership = await joinCommunity(area.id);
    communityStore.addCommunity({
      communityId: area.id,
      communityName: area.name,
      address: area.address,
    });
    savePendingJoin({
      communityId: area.id,
      communityName: area.name,
      address: area.address,
      membershipId: membership?.id,
    });
    uni.navigateTo({ url: '/pages/join-choice/join-choice' });
  } catch (e: any) {
    const msg = e?.message || e?.msg || '加入失败，请稍后重试';
    uni.showToast({ title: msg, icon: 'none', duration: 3000 });
  }
}

function isJoined(id: string): boolean {
  return communityStore.communities.some(c => c.communityId === id);
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; background: #FFFFFF; padding: 0 2rem; }

.header { padding: 1.875rem 0 1rem; text-align: center;
  .header-title { display: block; font-size: 1.375rem; font-weight: 700; color: $uni-text-color; margin-bottom: 0.25rem; }
  .header-sub { font-size: 0.8125rem; color: $uni-text-color-grey; }
}

.steps { display: flex; align-items: center; justify-content: center; margin-bottom: 1.25rem;
  .step { display: flex; flex-direction: column; align-items: center; gap: 0.1875rem;
    .step-num { width: 1.375rem; height: 1.375rem; border-radius: 50%; background: #F5F0EA; color: $uni-text-color-grey; font-size: 0.75rem; font-weight: 600; display: flex; align-items: center; justify-content: center; transition: all 0.3s; }
    .step-label { font-size: 0.625rem; color: $uni-text-color-grey; }
    &.active .step-num { background: $uni-color-primary; color: #fff; }
    &.active .step-label { color: $uni-color-primary; font-weight: 600; }
    &.done .step-num { background: #8DAF7E; color: #fff; }
    &.done .step-label { color: #8DAF7E; }
  }
  .step-line { width: 1.375rem; height: 0.0625rem; background: #E8E0D5; margin: 0 0.125rem 0.6875rem;
    &.done { background: #8DAF7E; }
  }
}

.card { background: #FAF8F5; border-radius: 0.5rem; padding: 1rem; box-shadow: $uni-shadow-base; }
.back-row { display: flex; align-items: center; gap: 0.25rem; margin-bottom: 0.625rem;
  .back-icon { font-size: 1rem; color: $uni-color-primary; }
  .back-text { font-size: 0.75rem; color: $uni-text-color-grey; }
}
.card-title { display: block; font-size: 0.9375rem; font-weight: 600; color: $uni-text-color; margin-bottom: 0.625rem; }
.loading-text, .empty-text { text-align: center; padding: 1.875rem 0; font-size: 0.8125rem; color: $uni-text-color-placeholder; }

.list { max-height: 16.25rem; }
.list-item { display: flex; align-items: center; justify-content: space-between; padding: 0.75rem 0; border-bottom: 0.03125rem solid #E8E0D5;
  &:last-child { border-bottom: none; }
  .item-content { flex: 1;
    .item-name { font-size: 0.875rem; color: $uni-text-color; }
    .item-addr { font-size: 0.6875rem; color: $uni-text-color-grey; margin-top: 0.125rem; display: block; }
  }
  .item-check { font-size: 1rem; color: $uni-color-primary; font-weight: 700; }
  .join-tag { padding: 0.25rem 0.625rem; border-radius: 0.75rem; background: $uni-color-primary; color: #fff; font-size: 0.6875rem; flex-shrink: 0; }
  .joined-text { font-size: 0.6875rem; color: #8DAF7E; flex-shrink: 0; }
  &.selected { background: rgba(184, 149, 106, 0.06); border-radius: 0.25rem; padding-left: 0.375rem; padding-right: 0.375rem; }
}

.btn { width: 100%; height: 2.75rem; border-radius: 1.375rem; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 0.9375rem; font-weight: 600; border: none; margin-top: 0.875rem;
  &::after { border: none; }
  &--disabled { background: #E8E0D5; color: #CCC4BA; }
}

.search-row { display: flex; gap: 0.5rem; margin-bottom: 0.75rem;
  .search-input { flex: 1; height: 2.5rem; background: #F5F0EA; border-radius: 0.375rem; padding: 0 0.625rem; font-size: 0.8125rem; color: $uni-text-color; }
  .search-btn { width: 3.75rem; height: 2.5rem; background: $uni-color-primary; border-radius: 0.375rem; display: flex; align-items: center; justify-content: center;
    text { color: #fff; font-size: 0.8125rem; font-weight: 600; }
  }
}

.max-banner { background: #FFF5F3; border-radius: 0.375rem; padding: 0.625rem 0.75rem; margin-top: 0.5rem;
  text { font-size: 0.75rem; color: #D4958A; line-height: 1.6; }
}
.max-inline { text-align: center; padding: 1.25rem 0;
  text { font-size: 0.875rem; color: #D4958A; }
}

.my-section-top { padding: 0.625rem 0; margin-bottom: 0.5rem;
  .my-title { font-size: 0.75rem; color: $uni-text-color-grey; margin-bottom: 0.375rem; display: block; }
  .my-tags { display: flex; flex-wrap: wrap; gap: 0.25rem; }
  .my-badge { display: inline-block; padding: 0.25rem 0.625rem; background: rgba(184, 149, 106, 0.1); border-radius: 0.25rem;
    text { font-size: 0.75rem; color: $uni-color-primary; }
  }
}

.max-warning { position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background: rgba(0,0,0,0.75); padding: 0.75rem 1.25rem; border-radius: 0.375rem;
  text { color: #fff; font-size: 0.875rem; }
}
</style>
