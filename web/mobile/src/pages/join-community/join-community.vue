<template>
  <view class="page">
    <!-- Header -->
    <view class="header">
      <text class="header-title">加入小区</text>
      <text class="header-sub">选择您的小区，开始社区生活</text>
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
      <view class="step" :class="{ active: step === 4, done: step > 4 }">
        <view class="step-num">{{ step > 4 ? '✓' : '4' }}</view>
        <text class="step-label">搜索小区</text>
      </view>
      <view class="step-line" :class="{ done: step > 4 }" />
      <view class="step" :class="{ active: step === 5 }">
        <view class="step-num">5</view>
        <text class="step-label">输入地址</text>
      </view>
    </view>

    <!-- Step 1: Select Province -->
    <view v-if="step === 1" class="card">
      <text class="card-title">请选择省份</text>
      <view v-if="provincesLoading" class="loading-text">加载中...</view>
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
      <button class="btn" :class="{ 'btn--disabled': !selectedProvince }" :disabled="!selectedProvince" @click="step = 2">
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

      <!-- My Communities -->
      <view v-if="myCommunities.length > 0" class="my-section">
        <text class="my-title">已加入 {{ myCommunities.length }}/3 个小区</text>
        <view v-for="c in myCommunities" :key="c.community_id" class="my-badge">
          <text>{{ c.communityName || '小区 ' + c.community_id }}</text>
        </view>
      </view>

      <!-- Search -->
      <view class="search-row">
        <input v-model="keyword" class="search-input" placeholder="输入小区名称模糊搜索" @confirm="doSearch" />
        <view class="search-btn" @click="doSearch"><text>搜索</text></view>
      </view>

      <!-- Results -->
      <view v-if="searching" class="loading-text">搜索中...</view>
      <scroll-view v-else-if="areas.length > 0" class="list" scroll-y>
        <view v-for="area in areas" :key="area.id" class="list-item" @click="goToStep5(area)">
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

    <!-- Step 5: Enter Address -->
    <view v-if="step === 5" class="card">
      <view class="back-row" @click="step = 4">
        <text class="back-icon">←</text>
        <text class="back-text">返回</text>
      </view>

      <!-- Selected Community Info -->
      <view class="community-info">
        <text class="community-icon">🏘️</text>
        <text class="community-name">{{ selectedCommunity?.name }}</text>
        <text v-if="selectedCommunity?.address" class="community-addr">{{ selectedCommunity?.address }}</text>
      </view>

      <!-- Address Format Example -->
      <view class="address-example">
        <text class="example-title">示例</text>
        <text class="example-text">5号楼 2单元 301房间：5-2-301</text>
      </view>

      <!-- Address Inputs -->
      <view class="address-inputs-row">
        <view class="input-col">
          <text class="input-label">楼号</text>
          <input
            v-model="step5building"
            class="addr-input"
            :class="{ 'input-error': buildingError }"
            type="number"
            placeholder="例：5"
          />
        </view>
        <text class="input-sep">-</text>
        <view class="input-col">
          <text class="input-label">单元号</text>
          <input
            v-model="step5unit"
            class="addr-input"
            :class="{ 'input-error': unitError }"
            type="number"
            placeholder="例：2"
          />
        </view>
        <text class="input-sep">-</text>
        <view class="input-col">
          <text class="input-label">房号</text>
          <input
            v-model="step5room"
            class="addr-input"
            :class="{ 'input-error': roomError }"
            type="number"
            placeholder="例：301"
          />
        </view>
      </view>

      <button class="btn" @click="submitJoin">确认加入</button>
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
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import {
  type Division, type ResidentialArea, type CommunityMembership,
  getDivisions, searchResidentialAreas,
  joinCommunity as joinCommunityApi, getUserMemberships,
} from '@/api/user';
import { useCommunityStore } from '@/stores/community';

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

// Memberships
const myCommunities = ref<(CommunityMembership & { communityName?: string })[]>([]);
const showMaxWarning = ref(false);

// Step 5: Enter Address
const selectedCommunity = ref<ResidentialArea | null>(null);
const step5building = ref('');
const step5unit = ref('');
const step5room = ref('');
const buildingError = ref(false);
const unitError = ref(false);
const roomError = ref(false);

onMounted(async () => {
  provincesLoading.value = true;
  try {
    const [divs, mems] = await Promise.all([
      getDivisions(),
      getUserMemberships().catch(() => [] as CommunityMembership[]),
    ]);
    provinces.value = divs;
    myCommunities.value = mems;
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

function goToStep5(area: ResidentialArea) {
  if (myCommunities.value.length >= 3) {
    showMaxWarning.value = true;
    setTimeout(() => { showMaxWarning.value = false; }, 2000);
    return;
  }
  if (isJoined(area.id)) return;
  selectedCommunity.value = area;
  step5building.value = '';
  step5unit.value = '';
  step5room.value = '';
  buildingError.value = false;
  unitError.value = false;
  roomError.value = false;
  step.value = 5;
}

function validateStep5(): string | null {
  buildingError.value = false;
  unitError.value = false;
  roomError.value = false;

  const building = step5building.value.trim();
  const unit = step5unit.value.trim();
  const room = step5room.value.trim();

  if (!building) {
    buildingError.value = true;
    return '请输入楼号';
  }
  const buildingNum = Number(building);
  if (isNaN(buildingNum) || !Number.isInteger(buildingNum) || buildingNum < 1 || buildingNum > 150) {
    buildingError.value = true;
    return '楼号必须为数字，且不大于150';
  }

  if (!unit) {
    unitError.value = true;
    return '请输入单元号';
  }
  const unitNum = Number(unit);
  if (isNaN(unitNum) || !Number.isInteger(unitNum) || unitNum < 1 || unitNum > 5) {
    unitError.value = true;
    return '单元号必须为数字，且不大于5';
  }

  if (!room) {
    roomError.value = true;
    return '请输入房号';
  }
  if (!/^\d{3}$/.test(room)) {
    roomError.value = true;
    return '房号必须为3位数字';
  }

  return null;
}

async function submitJoin() {
  const error = validateStep5();
  if (error) {
    uni.showToast({ title: error, icon: 'none', duration: 2000 });
    return;
  }

  if (!selectedCommunity.value) return;

  try {
    uni.showLoading({ title: '加入中...', mask: true });
    const mem = await joinCommunityApi({
      community_id: selectedCommunity.value.id,
      building: Number(step5building.value.trim()),
      unit: Number(step5unit.value.trim()),
      room: Number(step5room.value.trim()),
    });
    myCommunities.value.push({ ...mem, communityName: selectedCommunity.value.name });
    const communityStore = useCommunityStore();
    communityStore.addCommunity({
      communityId: selectedCommunity.value.id,
      communityName: selectedCommunity.value.name,
      address: selectedCommunity.value.address,
    });
    joinedArea.value = selectedCommunity.value;
    uni.hideLoading();
  } catch (_) {
    uni.hideLoading();
    uni.showToast({ title: '加入失败，请稍后重试', icon: 'none', duration: 2000 });
  }
}

function goHome() {
  uni.switchTab({ url: '/pages/notice/notice' });
}

function isJoined(id: string): boolean {
  return myCommunities.value.some(m => m.community_id === id);
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

.my-section { margin-bottom: 24rpx; padding: 20rpx; background: #fff; border-radius: 12rpx;
  .my-title { font-size: 24rpx; color: $uni-text-color-grey; margin-bottom: 12rpx; display: block; }
  .my-badge { display: inline-block; padding: 6rpx 16rpx; background: rgba(184, 149, 106, 0.1); border-radius: 8rpx; margin-right: 12rpx; margin-bottom: 8rpx;
    text { font-size: 22rpx; color: $uni-color-primary; }
  }
}

.max-warning { position: fixed; top: 50%; left: 50%; transform: translate(-50%, -50%); background: rgba(0,0,0,0.75); padding: 24rpx 40rpx; border-radius: 12rpx;
  text { color: #fff; font-size: 28rpx; }
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
// Step 5: Enter Address
.community-info { background: #fff; border-radius: 12rpx; padding: 28rpx; margin-bottom: 24rpx; text-align: center;
  .community-icon { font-size: 48rpx; display: block; margin-bottom: 8rpx; }
  .community-name { font-size: 32rpx; font-weight: 700; color: $uni-text-color; display: block; margin-bottom: 4rpx; }
  .community-addr { font-size: 24rpx; color: $uni-text-color-grey; display: block; }
}

.address-example { background: #fff; border-radius: 12rpx; padding: 20rpx 24rpx; margin-bottom: 32rpx;
  .example-title { font-size: 22rpx; color: $uni-text-color-grey; display: block; margin-bottom: 6rpx; }
  .example-text { font-size: 26rpx; color: $uni-text-color; display: block; }
}

.address-inputs-row { display: flex; align-items: flex-start; justify-content: center; gap: 12rpx; margin-bottom: 8rpx;
  .input-col { display: flex; flex-direction: column; align-items: center; flex: 1; max-width: 180rpx; }
  .input-label { font-size: 24rpx; color: $uni-text-color; margin-bottom: 10rpx; }
  .addr-input { width: 100%; height: 80rpx; background: #FAF8F5; border: 2rpx solid #E8DCCF; border-radius: 14rpx; text-align: center; font-size: 28rpx; color: $uni-text-color; padding: 0 8rpx;
    &.input-error { border-color: #D4958A; background: #FFF5F3; }
  }
  .input-sep { font-size: 36rpx; color: $uni-text-color-grey; font-weight: 600; line-height: 80rpx; padding-top: 34rpx; }
}
</style>
