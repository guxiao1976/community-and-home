<template>
  <view class="page">
    <!-- Loading state -->
    <view v-if="pageLoading" class="loading-wrap">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- Unauthenticated state -->
    <view v-else-if="!userStore.isLoggedIn" class="login-prompt" @click="goLogin">
      <view class="login-avatar">
        <text class="login-avatar-emoji">👤</text>
      </view>
      <text class="login-text">点击登录</text>
    </view>

    <!-- Authenticated: V9 Card Layout -->
    <template v-else>
      <!-- Gradient Header -->
      <view class="header">
        <view class="avatar-circle">
          <text class="avatar-emoji">👤</text>
        </view>
        <text class="header-name">{{ userStore.nickname || '当前用户' }}</text>
        <text class="header-phone">{{ displayPhone }}</text>
      </view>

      <!-- Community Management Section -->
      <view class="section">
        <view class="section-header">
          <text class="section-title">🏘️ 小区管理</text>
          <text class="section-badge">已加入 {{ communityStore.communityCount }}/3</text>
        </view>
        <view class="card-row">
          <view class="action-card" hover-class="action-card--hover" @click="goJoinCommunity">
            <view class="action-icon-box action-icon-box--join">
              <text class="action-icon-emoji">➕</text>
            </view>
            <text class="action-label">加入小区</text>
          </view>
          <view class="action-card" hover-class="action-card--hover" @click="goLeaveCommunity">
            <view class="action-icon-box action-icon-box--leave">
              <text class="action-icon-emoji">🚪</text>
            </view>
            <text class="action-label">退出小区</text>
          </view>
        </view>
      </view>

      <!-- Identity Section -->
      <view class="section">
        <view class="settings-box">
          <view class="settings-header">
            <text>🪪</text>
            <text class="settings-title">身份认证</text>
          </view>

          <!-- Owner (only if has communities) -->
          <view v-if="communityStore.hasCommunities" class="setting-item" @click="startOwnerAuth">
            <text>业主认证</text>
            <text class="item-hint">需绑房</text>
            <text class="arrow">→</text>
          </view>

          <!-- Tenant (only if has communities) -->
          <view v-if="communityStore.hasCommunities" class="setting-item" @click="startTenantAuth">
            <text>租户认证</text>
            <text class="item-hint">需绑房</text>
            <text class="arrow">→</text>
          </view>

          <!-- Committee (only if has approved owner role) -->
          <view v-if="hasOwnerRole" class="setting-item" @click="applyForRole('committee')">
            <text>业委会认证</text>
            <text class="item-hint">需先认证业主</text>
            <text class="arrow">→</text>
          </view>

          <!-- Always visible roles -->
          <view class="setting-item" @click="applyForRole('grid_worker')">
            <text>网格员认证</text>
            <text class="arrow">→</text>
          </view>
          <view class="setting-item" @click="applyForRole('property_admin')">
            <text>物业管理员认证</text>
            <text class="arrow">→</text>
          </view>
          <view class="setting-item" @click="applyForRole('community_admin')">
            <text>社区管理员认证</text>
            <text class="arrow">→</text>
          </view>
          <view class="setting-item setting-item--last" @click="applyForRole('merchant')">
            <text>商家认证</text>
            <text class="arrow">→</text>
          </view>
        </view>
      </view>

      <!-- Settings Section -->
      <view class="section">
        <view class="section-header">
          <text class="section-title">⚙️ 设置</text>
        </view>
        <view class="settings-list">
          <view class="settings-item" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">个人信息</text>
            <text class="settings-item-arrow">→</text>
          </view>
          <view class="settings-item" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">账号安全</text>
            <text class="settings-item-arrow">→</text>
          </view>
          <view class="settings-item settings-item--last" hover-class="settings-item--hover" @click="showDevToast">
            <text class="settings-item-label">关于我们</text>
            <text class="settings-item-arrow">→</text>
          </view>
        </view>
      </view>

      <!-- Bind Residence Modal -->
      <view v-if="showBindResidence" class="modal-mask" @click="showBindResidence = false">
        <view class="modal-box" @click.stop>
          <text class="modal-title">绑定房产</text>
          <text class="modal-sub">认证{{ authTarget === 'owner' ? '业主' : '租户' }}身份需要绑定房产</text>

          <!-- Example -->
          <view class="address-example">
            <text>示例：5号楼 2单元 301房间 → 5-2-301</text>
          </view>

          <!-- Three inputs -->
          <view class="address-inputs-row">
            <view class="input-col">
              <text class="input-label">楼号</text>
              <input v-model="bindBuilding" type="number" placeholder="如 5" class="addr-input" />
            </view>
            <text class="input-sep">-</text>
            <view class="input-col">
              <text class="input-label">单元号</text>
              <input v-model="bindUnit" type="number" placeholder="2" class="addr-input" />
            </view>
            <text class="input-sep">-</text>
            <view class="input-col">
              <text class="input-label">房号</text>
              <input v-model="bindRoom" type="number" placeholder="301" class="addr-input" />
            </view>
          </view>

          <view class="modal-btns">
            <view class="btn-cancel" @click="showBindResidence = false">取消</view>
            <view class="btn-confirm" @click="submitBindAndApply">确认绑定并申请</view>
          </view>
        </view>
      </view>
    </template>
  </view>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';
import { applyRole, bindResidence, getUserMemberships } from '@/api/user';

const userStore = useUserStore();
const communityStore = useCommunityStore();

const pageLoading = ref(true);

// Identity / role state
const showBindResidence = ref(false);
const authTarget = ref<'owner' | 'tenant'>('owner');
const bindBuilding = ref('');
const bindUnit = ref('');
const bindRoom = ref('');

const hasOwnerRole = computed(() => {
  // Placeholder — will be enhanced with API-loaded roles in future iteration
  return false;
});

// Phone display with masking
const displayPhone = computed(() => {
  try {
    const phone = uni.getStorageSync('user_phone') as string;
    if (phone && phone.length >= 11) {
      return phone.slice(0, 3) + '****' + phone.slice(-4);
    }
  } catch {
    // ignore storage errors
  }
  return '未绑定手机号';
});

// Navigation
function goLogin() {
  uni.navigateTo({ url: '/pages/login/login' });
}

function goJoinCommunity() {
  uni.navigateTo({ url: '/pages/join-community/join-community' });
}

function goLeaveCommunity() {
  uni.navigateTo({ url: '/pages/leave-community/leave-community' });
}

function showDevToast() {
  uni.showToast({ title: '页面开发中', icon: 'none', duration: 1500 });
}

// Identity / role actions
function startOwnerAuth() {
  authTarget.value = 'owner';
  showBindResidence.value = true;
}

function startTenantAuth() {
  authTarget.value = 'tenant';
  showBindResidence.value = true;
}

async function submitBindAndApply() {
  const b = bindBuilding.value.trim();
  const u = bindUnit.value.trim();
  const r = bindRoom.value.trim();

  if (!b || !u || !r) {
    uni.showToast({ title: '请填写完整的楼号、单元号、房号', icon: 'none' });
    return;
  }

  const currentCommunityId = communityStore.currentCommunityId;
  if (!currentCommunityId) {
    uni.showToast({ title: '请先加入小区', icon: 'none' });
    return;
  }

  try {
    // Find the membership_id for the current community
    const memberships = await getUserMemberships();
    const membership = memberships.find(
      (m: any) => (m.community_id || m.communityId) === currentCommunityId,
    );
    if (!membership) {
      uni.showToast({ title: '未找到小区成员关系', icon: 'none' });
      return;
    }
    const membershipId = membership.id;

    // Step 1: Bind residence
    await bindResidence({
      membership_id: membershipId,
      building: b,
      unit: u,
      room: r,
      is_primary: 1,
    });

    // Step 2: Apply for role
    await applyRole({
      community_id: currentCommunityId,
      role_code: authTarget.value,
    });

    showBindResidence.value = false;
    uni.showToast({ title: '认证申请已提交，等待审核', icon: 'success' });
  } catch (e: any) {
    uni.showToast({ title: e.message || '操作失败', icon: 'none' });
  }
}

function applyForRole(roleCode: string) {
  // TODO: In production, this should call applyRole API with proper community_id
  // For non-owner/tenant roles (no bind-residence needed), submit directly
  const currentCommunityId = communityStore.currentCommunityId;
  if (!currentCommunityId) {
    uni.showToast({ title: '请先加入小区', icon: 'none' });
    return;
  }
  applyRole({
    community_id: currentCommunityId,
    role_code: roleCode,
  }).then(() => {
    uni.showToast({ title: '认证申请已提交，等待审核', icon: 'success' });
  }).catch((e: any) => {
    uni.showToast({ title: e.message || '操作失败', icon: 'none' });
  });
}

onMounted(async () => {
  // Ensure user profile is loaded (token may exist without user in store)
  if (isAuthenticated() && !userStore.user) {
    try {
      const user = await getUserProfile();
      userStore.setUser(user);
    } catch (e) {
      console.warn('[my] Failed to load user profile:', e);
    }
  }
  if (userStore.isLoggedIn) {
    await communityStore.loadMemberships();
  }
  pageLoading.value = false;
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: #FFFFFF;
}

// ---- Loading ----
.loading-wrap {
  display: flex;
  justify-content: center;
  padding: 200rpx 0;
}

.loading-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}

// ---- Login prompt ----
.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 200rpx 0;
}

.login-avatar {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 3rpx solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 4rpx 20rpx rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24rpx;
}

.login-avatar-emoji {
  font-size: 72rpx;
}

.login-text {
  font-size: 32rpx;
  color: #B8956A;
}

// ---- Gradient Header ----
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 0 60rpx;
  background: linear-gradient(160deg, #D4B896 0%, #E8DCCF 30%, #FAF8F5 55%, #FFFFFF 80%);
}

.avatar-circle {
  width: 140rpx;
  height: 140rpx;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 3rpx solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 4rpx 20rpx rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20rpx;
}

.avatar-emoji {
  font-size: 72rpx;
}

.header-name {
  font-size: 24rpx;
  color: #A6988A;
  margin-bottom: 8rpx;
}

.header-phone {
  font-size: 32rpx;
  font-weight: 600;
  color: #3D3226;
}

// ---- Sections ----
.section {
  padding: 0 40rpx;
  margin-top: 32rpx;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #3D3226;
}

.section-badge {
  font-size: 24rpx;
  color: #B8956A;
  background-color: rgba(184, 149, 106, 0.1);
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
}

// ---- Action Cards ----
.card-row {
  display: flex;
  gap: 20rpx;
}

.action-card {
  flex: 1;
  height: 260rpx;
  background-color: #FAF8F5;
  border-radius: 24rpx;
  box-shadow: 0 4rpx 16rpx rgba(184, 149, 106, 0.08);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;

  &--hover {
    opacity: 0.85;
    transform: scale(0.97);
    transition: all 0.15s ease;
  }
}

.action-icon-box {
  width: 160rpx;
  height: 160rpx;
  border-radius: 32rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 16rpx;

  &--join {
    background: linear-gradient(135deg, #B8956A, #D4B896);
  }

  &--leave {
    background: linear-gradient(135deg, #D4958A, #E0ADA5);
  }
}

.action-icon-emoji {
  font-size: 84rpx;
}

.action-label {
  font-size: 28rpx;
  font-weight: 500;
  color: #3D3226;
}

// ---- Settings ----
.settings-list {
  background-color: #FAF8F5;
  border-radius: 20rpx;
  padding: 28rpx 32rpx;
}

.settings-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14rpx 0;
  border-bottom: 1rpx solid rgba(184, 149, 106, 0.1);

  &--last {
    border-bottom: none;
  }

  &--hover {
    background-color: rgba(184, 149, 106, 0.04);
    transition: background-color 0.15s ease;
  }
}

.settings-item-label {
  font-size: 28rpx;
  color: #3D3226;
}

.settings-item-arrow {
  font-size: 24rpx;
  color: #CCC4BA;
}

// ---- Settings Header (for identity section) ----
.settings-box {
  background-color: #FAF8F5;
  border-radius: 20rpx;
  padding: 28rpx 32rpx;
}

.settings-header {
  display: flex;
  align-items: center;
  gap: 10rpx;
  margin-bottom: 12rpx;
}

.settings-title {
  font-size: 30rpx;
  font-weight: 600;
  color: #3D3226;
}

// ---- Identity hints ----
.item-hint {
  font-size: 20rpx;
  color: #B8956A;
  background: rgba(184, 149, 106, 0.08);
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
  margin-left: auto;
  margin-right: 8rpx;
}

.arrow {
  font-size: 24rpx;
  color: #CCC4BA;
}

// ---- Bind Residence Modal ----
.modal-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.modal-box {
  width: 640rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 40rpx 32rpx;
}

.modal-title {
  font-size: 34rpx;
  font-weight: 700;
  color: #3D3226;
  display: block;
  text-align: center;
  margin-bottom: 8rpx;
}

.modal-sub {
  font-size: 24rpx;
  color: #A6988A;
  display: block;
  text-align: center;
  margin-bottom: 24rpx;
}

.address-example {
  background: #FAF8F5;
  border-radius: 10rpx;
  padding: 16rpx 20rpx;
  margin-bottom: 24rpx;
  text-align: center;
}

.address-example text {
  font-size: 24rpx;
  color: #8C7B6B;
}

.address-inputs-row {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 12rpx;
  margin-bottom: 32rpx;
}

.input-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.input-label {
  font-size: 22rpx;
  color: #8C7B6B;
  margin-bottom: 8rpx;
}

.addr-input {
  width: 100%;
  height: 72rpx;
  background: #FAF8F5;
  border: 2rpx solid #E8DCCF;
  border-radius: 12rpx;
  text-align: center;
  font-size: 28rpx;
  color: #3D3226;
  padding: 0 8rpx;
}

.input-sep {
  font-size: 32rpx;
  color: #CCC4BA;
  line-height: 72rpx;
  padding-top: 30rpx;
}

.modal-btns {
  display: flex;
  gap: 16rpx;
}

.btn-cancel {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  background: #F0EBE3;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #8C7B6B;
}

.btn-confirm {
  flex: 1;
  height: 80rpx;
  border-radius: 40rpx;
  background: linear-gradient(135deg, #B8956A, #D4B896);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  color: #fff;
  font-weight: 600;
}
</style>
