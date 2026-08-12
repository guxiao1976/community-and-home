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
      <button class="btn" :class="{ 'btn--disabled': !selectedProvince || maxReached }" :disabled="!selectedProvince || maxReached" @click="step = 2">
        下一步
      </button>
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
      <button class="btn" :class="{ 'btn--disabled': !selectedCity }" :disabled="!selectedCity" @click="step = 3">
        下一步
      </button>
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
        <view v-for="area in areas" :key="area.id" class="list-item" @click="openJoinForm(area)">
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

    <!-- Join Success -->
    <view v-if="joinedArea" class="success-card">
      <view class="success-icon">🎉</view>
      <text class="success-title">加入成功！</text>
      <text class="success-name">{{ joinedArea.name }}</text>
      <text v-if="joinedArea.address" class="success-addr">{{ joinedArea.address }}</text>
      <view class="success-actions">
        <view class="success-btn" @click="goHome">回首页</view>
        <view class="success-btn-outline" @click="joinedArea = null">继续加入</view>
      </view>
    </view>

    <!-- Max Warning -->
    <view v-if="showMaxWarning" class="max-warning"><text>最多加入 3 个小区</text></view>

    <!-- Join Form Modal: ownership (自有/租住, required) + building/unit/room -->
    <view v-if="showJoinForm && joinTarget" class="join-form-mask" @click.self="closeJoinForm">
      <view class="join-form-card">
        <view class="join-form-title">加入 {{ joinTarget.name }}</view>

        <view class="join-form-label">房屋权属 <text class="required">*</text></view>
        <view class="ownership-row">
          <view
            v-for="opt in OWNERSHIP_OPTIONS"
            :key="opt.value"
            class="ownership-option"
            :class="{ selected: joinForm.ownership === opt.value }"
            @click="joinForm.ownership = opt.value"
          >
            {{ opt.label }}
          </view>
        </view>
        <text v-if="joinFormErrors.ownership" class="field-error">{{ joinFormErrors.ownership }}</text>

        <view class="join-form-label">楼号 <text class="required">*</text></view>
        <input class="join-form-input" v-model="joinForm.building" type="number" placeholder="如 3" />
        <text v-if="joinFormErrors.building" class="field-error">{{ joinFormErrors.building }}</text>

        <view class="join-form-label">单元号 <text class="required">*</text></view>
        <input class="join-form-input" v-model="joinForm.unit" type="number" placeholder="如 1" />
        <text v-if="joinFormErrors.unit" class="field-error">{{ joinFormErrors.unit }}</text>

        <view class="join-form-label">房号 <text class="required">*</text></view>
        <input class="join-form-input" v-model="joinForm.room" type="number" placeholder="如 502" />
        <text v-if="joinFormErrors.room" class="field-error">{{ joinFormErrors.room }}</text>

        <view class="join-form-actions">
          <view class="join-form-cancel" @click="closeJoinForm">取消</view>
          <view class="confirm-join-btn" @click="confirmJoin">确认加入</view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import {
  type Division, type ResidentialArea,
  getDivisions, searchResidentialAreas,
  joinCommunity,
} from '@/api/user';
import { useCommunityStore } from '@/stores/community';
import {
  OWNERSHIP_OPTIONS, validateJoinForm, joinFormToPayload,
  type JoinFormErrors, type JoinFormState,
} from './join-form';

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
}

async function selectCity(c: Division) {
  selectedCity.value = c;
  selectedDistrict.value = null;
  districtsLoading.value = true;
  try { districts.value = await getDivisions(c.id); } catch (_) { districts.value = []; }
  finally { districtsLoading.value = false; }
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

const joinedArea = ref<ResidentialArea | null>(null);

// --- Join form (ownership + building/unit/room) ---
const showJoinForm = ref(false);
const joinTarget = ref<ResidentialArea | null>(null);
const joinForm = ref<JoinFormState>({
  building: '',
  unit: '',
  room: '',
  ownership: null,
});
const joinFormErrors = ref<JoinFormErrors>({});

function openJoinForm(area: ResidentialArea) {
  if (maxReached.value) {
    showMaxWarning.value = true;
    setTimeout(() => { showMaxWarning.value = false; }, 2000);
    return;
  }
  if (isJoined(area.id)) return;
  joinTarget.value = area;
  joinForm.value = { building: '', unit: '', room: '', ownership: null };
  joinFormErrors.value = {};
  showJoinForm.value = true;
}

function closeJoinForm() {
  showJoinForm.value = false;
  joinTarget.value = null;
}

async function confirmJoin() {
  const target = joinTarget.value;
  if (!target) return;

  // 加入前收集「自有/租住」选择（必填）+ 楼/单元/房号输入
  const result = validateJoinForm(joinForm.value);
  joinFormErrors.value = result.errors;
  if (!result.valid) return;

  try {
    uni.hideLoading(); // clear any lingering loading state
    uni.showLoading({ title: '加入中...', mask: true });
    const { building, unit, room, ownership } = joinFormToPayload(joinForm.value);
    await joinCommunity(target.id, building, unit, room, ownership);
    communityStore.addCommunity({
      communityId: target.id,
      communityName: target.name,
      address: target.address,
    });
    showJoinForm.value = false;
    joinedArea.value = target;
    joinTarget.value = null;
  } catch (e: any) {
    const msg = e?.message || e?.msg || '加入失败，请稍后重试';
    uni.showToast({ title: msg, icon: 'none', duration: 3000 });
  } finally {
    uni.hideLoading();
  }
}

function goHome() {
  uni.switchTab({ url: '/pages/notice/notice' });
}

function isJoined(id: string): boolean {
  return communityStore.communities.some(c => c.communityId === id);
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; background: #FFFFFF; padding: 0 32px; }

.header { padding: 60rpx 0 32rpx; text-align: center;
  .header-title { display: block; font-size: 44rpx; font-weight: 700; color: $uni-text-color; margin-bottom: 8rpx; }
  .header-sub { font-size: 26rpx; color: $uni-text-color-grey; }
}

.steps { display: flex; align-items: center; justify-content: center; margin-bottom: 40rpx;
  .step { display: flex; flex-direction: column; align-items: center; gap: 6rpx;
    .step-num { width: 44rpx; height: 44rpx; border-radius: 50%; background: #F5F0EA; color: $uni-text-color-grey; font-size: 24rpx; font-weight: 600; display: flex; align-items: center; justify-content: center; transition: all 0.3s; }
    .step-label { font-size: 20rpx; color: $uni-text-color-grey; }
    &.active .step-num { background: $uni-color-primary; color: #fff; }
    &.active .step-label { color: $uni-color-primary; font-weight: 600; }
    &.done .step-num { background: #8DAF7E; color: #fff; }
    &.done .step-label { color: #8DAF7E; }
  }
  .step-line { width: 44rpx; height: 2rpx; background: #E8E0D5; margin: 0 4rpx 22rpx;
    &.done { background: #8DAF7E; }
  }
}

.card { background: #FAF8F5; border-radius: 16rpx; padding: 32rpx; box-shadow: $uni-shadow-base; }
.back-row { display: flex; align-items: center; gap: 8rpx; margin-bottom: 20rpx;
  .back-icon { font-size: 32rpx; color: $uni-color-primary; }
  .back-text { font-size: 24rpx; color: $uni-text-color-grey; }
}
.card-title { display: block; font-size: 30rpx; font-weight: 600; color: $uni-text-color; margin-bottom: 20rpx; }
.loading-text, .empty-text { text-align: center; padding: 60rpx 0; font-size: 26rpx; color: $uni-text-color-placeholder; }

.list { max-height: 520rpx; }
.list-item { display: flex; align-items: center; justify-content: space-between; padding: 24rpx 0; border-bottom: 1rpx solid #E8E0D5;
  &:last-child { border-bottom: none; }
  .item-content { flex: 1;
    .item-name { font-size: 28rpx; color: $uni-text-color; }
    .item-addr { font-size: 22rpx; color: $uni-text-color-grey; margin-top: 4rpx; display: block; }
  }
  .item-check { font-size: 32rpx; color: $uni-color-primary; font-weight: 700; }
  .join-tag { padding: 8rpx 20rpx; border-radius: 24rpx; background: $uni-color-primary; color: #fff; font-size: 22rpx; flex-shrink: 0; }
  .joined-text { font-size: 22rpx; color: #8DAF7E; flex-shrink: 0; }
  &.selected { background: rgba(184, 149, 106, 0.06); border-radius: 8rpx; padding-left: 12rpx; padding-right: 12rpx; }
}

.btn { width: 100%; height: 88rpx; border-radius: 44rpx; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 30rpx; font-weight: 600; border: none; margin-top: 28rpx;
  &::after { border: none; }
  &--disabled { background: #E8E0D5; color: #CCC4BA; }
}

.search-row { display: flex; gap: 16rpx; margin-bottom: 24rpx;
  .search-input { flex: 1; height: 80rpx; background: #F5F0EA; border-radius: 12rpx; padding: 0 20rpx; font-size: 26rpx; color: $uni-text-color; }
  .search-btn { width: 120rpx; height: 80rpx; background: $uni-color-primary; border-radius: 12rpx; display: flex; align-items: center; justify-content: center;
    text { color: #fff; font-size: 26rpx; font-weight: 600; }
  }
}

.max-banner { background: #FFF5F3; border-radius: 12rpx; padding: 20rpx 24rpx; margin-top: 16rpx;
  text { font-size: 24rpx; color: #D4958A; line-height: 1.6; }
}
.max-inline { text-align: center; padding: 40rpx 0;
  text { font-size: 28rpx; color: #D4958A; }
}

.my-section-top { padding: 20rpx 0; margin-bottom: 16rpx;
  .my-title { font-size: 24rpx; color: $uni-text-color-grey; margin-bottom: 12rpx; display: block; }
  .my-tags { display: flex; flex-wrap: wrap; gap: 8rpx; }
  .my-badge { display: inline-block; padding: 8rpx 20rpx; background: rgba(184, 149, 106, 0.1); border-radius: 8rpx;
    text { font-size: 24rpx; color: $uni-color-primary; }
  }
}

.max-warning { position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background: rgba(0,0,0,0.75); padding: 24rpx 40rpx; border-radius: 12rpx;
  text { color: #fff; font-size: 28rpx; }
}

.join-form-mask { position: fixed; inset: 0; background: rgba(0,0,0,0.5); z-index: 100; display: flex; align-items: center; justify-content: center; padding: 0 40rpx; }
.join-form-card { background: #FFFFFF; border-radius: 24rpx; padding: 40rpx 32rpx; width: 100%; max-width: 600rpx; box-shadow: $uni-shadow-base; }
.join-form-title { font-size: 34rpx; font-weight: 700; color: $uni-text-color; text-align: center; margin-bottom: 32rpx; }
.join-form-label { font-size: 26rpx; color: $uni-text-color; margin-bottom: 12rpx;
  .required { color: $uni-color-error; }
}
.ownership-row { display: flex; gap: 20rpx; margin-bottom: 8rpx;
  .ownership-option { flex: 1; text-align: center; padding: 20rpx 0; border-radius: 12rpx; background: $uni-bg-color-input; color: $uni-text-color-grey; font-size: 28rpx; border: 2rpx solid transparent;
    &.selected { background: rgba(184, 149, 106, 0.12); color: $uni-color-primary; border-color: $uni-color-primary; font-weight: 600; }
  }
}
.join-form-input { width: 100%; height: 80rpx; background: $uni-bg-color-input; border-radius: 12rpx; padding: 0 20rpx; font-size: 28rpx; color: $uni-text-color; box-sizing: border-box; margin-bottom: 8rpx; }
.field-error { display: block; font-size: 22rpx; color: $uni-color-error; margin-bottom: 16rpx; }
.join-form-actions { display: flex; gap: 24rpx; margin-top: 32rpx;
  .join-form-cancel { flex: 1; text-align: center; padding: 20rpx 0; border-radius: 44rpx; border: 2rpx solid #E8E0D5; color: $uni-text-color-grey; font-size: 28rpx; }
  .confirm-join-btn { flex: 1; text-align: center; padding: 20rpx 0; border-radius: 44rpx; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 28rpx; font-weight: 600; }
}

.success-card { background: #FAF8F5; border-radius: 16rpx; padding: 48rpx 32rpx; text-align: center; box-shadow: $uni-shadow-base; margin-top: 32rpx;
  .success-icon { font-size: 72rpx; margin-bottom: 16rpx; }
  .success-title { display: block; font-size: 36rpx; font-weight: 700; color: $uni-text-color; margin-bottom: 8rpx; }
  .success-name { display: block; font-size: 28rpx; color: $uni-color-primary; font-weight: 600; margin-bottom: 4rpx; }
  .success-addr { display: block; font-size: 22rpx; color: $uni-text-color-grey; margin-bottom: 32rpx; }
  .success-actions { display: flex; gap: 24rpx; justify-content: center; }
  .success-btn { padding: 16rpx 48rpx; border-radius: 44rpx; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 28rpx; font-weight: 600; }
  .success-btn-outline { padding: 16rpx 48rpx; border-radius: 44rpx; border: 2rpx solid #B8956A; color: #B8956A; font-size: 28rpx; font-weight: 600; }
}
</style>
