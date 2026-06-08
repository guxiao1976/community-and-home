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

    <!-- Authenticated: Hierarchical List Layout -->
    <template v-else>
      <!-- Gradient Header -->
      <view class="header">
        <view class="avatar-circle">
          <text class="avatar-emoji">👤</text>
        </view>
        <text class="header-name">{{ userStore.nickname || '当前用户' }}</text>
        <text class="header-phone">{{ displayPhone }}</text>
      </view>

      <!-- 1. 小区管理 -->
      <view class="menu-section">
        <view class="menu-row" hover-class="menu-row--hover" @click="expanded = (expanded === 'community' ? '' : 'community')">
          <view class="menu-left">
            <text class="menu-icon">🏘️</text>
            <view class="menu-text">
              <text class="menu-title">小区管理</text>
              <text class="menu-desc">加入小区 · 退出小区 · 已加入 {{ communityStore.communityCount }}/3</text>
            </view>
          </view>
          <text class="menu-arrow" :class="{ 'menu-arrow--open': expanded === 'community' }">→</text>
        </view>
        <view v-if="expanded === 'community'" class="sub-menu">
          <view class="sub-item" hover-class="sub-item--hover" @click="goJoinCommunity">
            <view class="sub-icon sub-icon--join">➕</view>
            <text class="sub-label">加入小区</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="goLeaveCommunity">
            <view class="sub-icon sub-icon--leave">🚪</view>
            <text class="sub-label">退出小区</text>
            <text class="sub-arrow">→</text>
          </view>
        </view>
      </view>

      <!-- 2. 业主/租户登记 -->
      <view class="menu-section">
        <view class="menu-row" hover-class="menu-row--hover" @click="expanded = (expanded === 'registration' ? '' : 'registration')">
          <view class="menu-left">
            <text class="menu-icon">🏠</text>
            <view class="menu-text">
              <text class="menu-title">业主/租户登记</text>
              <text class="menu-desc">业主登记 · 租户登记</text>
            </view>
          </view>
          <text class="menu-arrow" :class="{ 'menu-arrow--open': expanded === 'registration' }">→</text>
        </view>
        <view v-if="expanded === 'registration'" class="sub-menu">
          <view v-if="communityStore.hasCommunities" class="sub-item" hover-class="sub-item--hover" @click="startOwnerAuth">
            <text class="sub-label">业主登记</text>
            <text class="item-hint">登记房号</text>
            <text class="sub-arrow">→</text>
          </view>
          <view v-if="communityStore.hasCommunities" class="sub-item" hover-class="sub-item--hover" @click="startTenantAuth">
            <text class="sub-label">租户登记</text>
            <text class="item-hint">登记房号</text>
            <text class="sub-arrow">→</text>
          </view>
          <view v-if="!communityStore.hasCommunities" class="sub-item sub-item--disabled">
            <text class="sub-label sub-label--muted">请先加入小区</text>
          </view>
        </view>
      </view>

      <!-- 3. 身份认证 -->
      <view class="menu-section">
        <view class="menu-row" hover-class="menu-row--hover" @click="expanded = (expanded === 'identity' ? '' : 'identity')">
          <view class="menu-left">
            <text class="menu-icon">🪪</text>
            <view class="menu-text">
              <text class="menu-title">身份认证</text>
              <text class="menu-desc">业委会 · 网格员 · 物业管理员 · 社区管理员 · 商家</text>
            </view>
          </view>
          <text class="menu-arrow" :class="{ 'menu-arrow--open': expanded === 'identity' }">→</text>
        </view>
        <view v-if="expanded === 'identity'" class="sub-menu">
          <view v-if="hasOwnerRole" class="sub-item" hover-class="sub-item--hover" @click="applyForRole('committee')">
            <text class="sub-label">业委会认证</text>
            <text class="item-hint">需先认证业主</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="applyForRole('grid_worker')">
            <text class="sub-label">网格员认证</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="applyForRole('property_admin')">
            <text class="sub-label">物业管理员认证</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="applyForRole('community_admin')">
            <text class="sub-label">社区管理员认证</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="applyForRole('merchant')">
            <text class="sub-label">商家认证</text>
            <text class="sub-arrow">→</text>
          </view>
        </view>
      </view>

      <!-- 4. 账户管理 -->
      <view class="menu-section">
        <view class="menu-row" hover-class="menu-row--hover" @click="expanded = (expanded === 'account' ? '' : 'account')">
          <view class="menu-left">
            <text class="menu-icon">⚙️</text>
            <view class="menu-text">
              <text class="menu-title">账户管理</text>
              <text class="menu-desc">个人信息 · 账号安全 · 关于我们</text>
            </view>
          </view>
          <text class="menu-arrow" :class="{ 'menu-arrow--open': expanded === 'account' }">→</text>
        </view>
        <view v-if="expanded === 'account'" class="sub-menu">
          <view class="sub-item" hover-class="sub-item--hover" @click="showDevToast">
            <text class="sub-label">个人信息</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="goAccountSecurity">
            <text class="sub-label">账号安全</text>
            <text class="sub-arrow">→</text>
          </view>
          <view class="sub-item" hover-class="sub-item--hover" @click="showDevToast">
            <text class="sub-label">关于我们</text>
            <text class="sub-arrow">→</text>
          </view>
        </view>
      </view>

      <!-- Bind Residence Modal -->
      <view v-if="showBindResidence" class="modal-mask" @click="showBindResidence = false">
        <view class="modal-box" @click.stop>
          <text class="modal-title">{{ authTarget === 'owner' ? '业主' : '租户' }}登记</text>

          <!-- Step 1: Select Community -->
          <text class="modal-label">选择小区</text>
          <scroll-view v-if="communityStore.communities.length > 1" class="community-picker" scroll-y>
            <view
              v-for="c in communityStore.communities"
              :key="c.communityId"
              class="community-option"
              :class="{ 'community-option--selected': bindCommunityId === c.communityId }"
              @click="bindCommunityId = c.communityId"
            >
              <text>{{ c.communityName }}</text>
              <text v-if="bindCommunityId === c.communityId" class="community-check">✓</text>
            </view>
          </scroll-view>
          <view v-else class="community-single">
            <text>{{ communityStore.communities[0]?.communityName || '当前小区' }}</text>
          </view>

          <text class="modal-label" style="margin-top: 24rpx;">登记房号</text>
          <view class="address-example">
            <text>示例：5号楼 2单元 301房间 → 5-2-301</text>
          </view>
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
            <view class="btn-confirm" @click="submitBindAndApply">确认登记</view>
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

// Hierarchical menu expansion state: '' | 'community' | 'identity' | 'account'
const expanded = ref('');

// Identity / role state
const showBindResidence = ref(false);
const authTarget = ref<'owner' | 'tenant'>('owner');
const bindCommunityId = ref('');
const bindBuilding = ref('');
const bindUnit = ref('');
const bindRoom = ref('');

const hasOwnerRole = computed(() => {
  // Placeholder — will be enhanced with API-loaded roles in future iteration
  return false;
});

// Phone display with masking (read from store first, fallback to storage)
const displayPhone = computed(() => {
  const phone = userStore.user?.phone;
  if (phone && phone.length >= 11) {
    return phone.slice(0, 3) + '****' + phone.slice(-4);
  }
  try {
    const stored = uni.getStorageSync('user_phone') as string;
    if (stored && stored.length >= 11) {
      return stored.slice(0, 3) + '****' + stored.slice(-4);
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

function goAccountSecurity() {
  uni.navigateTo({ url: '/pages/account-security/account-security' });
}

function showDevToast() {
  uni.showToast({ title: '页面开发中', icon: 'none', duration: 1500 });
}

// Identity / role actions
function startOwnerAuth() {
  authTarget.value = 'owner';
  bindCommunityId.value = communityStore.currentCommunityId || communityStore.communities[0]?.communityId || '';
  bindBuilding.value = '';
  bindUnit.value = '';
  bindRoom.value = '';
  showBindResidence.value = true;
}

function startTenantAuth() {
  authTarget.value = 'tenant';
  bindCommunityId.value = communityStore.currentCommunityId || communityStore.communities[0]?.communityId || '';
  bindBuilding.value = '';
  bindUnit.value = '';
  bindRoom.value = '';
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

  const targetCommunityId = bindCommunityId.value;
  if (!targetCommunityId) {
    uni.showToast({ title: '请先选择小区', icon: 'none' });
    return;
  }

  try {
    // Find the membership_id for the selected community
    const memberships = await getUserMemberships();
    const membership = memberships.find(
      (m: any) => (m.community_id || m.communityId) === targetCommunityId,
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
      community_id: targetCommunityId,
      role_code: authTarget.value,
    });

    showBindResidence.value = false;
    uni.showToast({ title: '房号登记成功', icon: 'success' });
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

// ---- Hierarchical Menu Sections ----
.menu-section {
  padding: 0 40rpx;
  margin-top: 20rpx;
}

.menu-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  background-color: #FAF8F5;
  border-radius: 16rpx;
  padding: 32rpx;

  &--hover {
    background-color: rgba(184, 149, 106, 0.06);
    transition: background-color 0.15s ease;
  }
}

.menu-left {
  display: flex;
  align-items: center;
  gap: 16rpx;
  flex: 1;
  min-width: 0;
}

.menu-icon {
  font-size: 44rpx;
  flex-shrink: 0;
}

.menu-text {
  flex: 1;
  min-width: 0;
}

.menu-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #3D3226;
  display: block;
}

.menu-desc {
  font-size: 24rpx;
  color: #A6988A;
  margin-top: 6rpx;
  display: block;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.menu-arrow {
  font-size: 28rpx;
  color: #CCC4BA;
  flex-shrink: 0;
  transition: transform 0.2s ease;

  &--open {
    transform: rotate(90deg);
  }
}

// ---- Sub-menu (expanded items) ----
.sub-menu {
  padding: 12rpx 0 0 60rpx;
}

.sub-item {
  display: flex;
  align-items: center;
  padding: 20rpx 24rpx;
  background-color: #FAF8F5;
  border-radius: 12rpx;
  margin-bottom: 8rpx;

  &:last-child {
    margin-bottom: 0;
  }

  &--hover {
    background-color: rgba(184, 149, 106, 0.08);
    transition: background-color 0.15s ease;
  }
}

.sub-icon {
  width: 56rpx;
  height: 56rpx;
  border-radius: 14rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30rpx;
  color: #fff;
  margin-right: 16rpx;
  flex-shrink: 0;

  &--join {
    background: linear-gradient(135deg, #B8956A, #D4B896);
  }

  &--leave {
    background: linear-gradient(135deg, #D4958A, #E0ADA5);
  }
}

.sub-label {
  font-size: 28rpx;
  color: #3D3226;
  flex: 1;
}

.sub-arrow {
  font-size: 24rpx;
  color: #CCC4BA;
  flex-shrink: 0;
}

.sub-item--disabled {
  opacity: 0.5;
}

.sub-label--muted {
  font-size: 26rpx;
  color: #A6988A;
}

// ---- Identity hints ----
.item-hint {
  font-size: 20rpx;
  color: #B8956A;
  background: rgba(184, 149, 106, 0.08);
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
  margin-right: 8rpx;
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

.modal-label {
  font-size: 26rpx;
  font-weight: 600;
  color: #3D3226;
  display: block;
  margin-bottom: 16rpx;
}

// ---- Community Picker ----
.community-picker {
  max-height: 240rpx;
  margin-bottom: 8rpx;
}

.community-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16rpx 20rpx;
  background: #FAF8F5;
  border-radius: 10rpx;
  margin-bottom: 8rpx;
  font-size: 26rpx;
  color: #3D3226;
  border: 2rpx solid transparent;

  &--selected {
    border-color: #B8956A;
    background: rgba(184, 149, 106, 0.06);
  }
}

.community-check {
  font-size: 28rpx;
  color: #B8956A;
  font-weight: 700;
}

.community-single {
  padding: 16rpx 20rpx;
  background: #FAF8F5;
  border-radius: 10rpx;
  font-size: 26rpx;
  color: #3D3226;
  text-align: center;
  margin-bottom: 8rpx;
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
