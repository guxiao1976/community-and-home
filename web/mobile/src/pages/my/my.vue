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

    <!-- Authenticated: Icon Grid Layout -->
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
        <view class="section-header">
          <text class="section-icon">🏘️</text>
          <text class="section-title">小区管理</text>
        </view>
        <view class="func-entries">
          <view class="func-entry" hover-class="func-entry--hover" @click="goJoinCommunity">
            <text class="func-entry-icon">🏘️</text>
            <text class="func-entry-label">加入小区</text>
          </view>
          <view class="func-entry" hover-class="func-entry--hover" @click="goLeaveCommunity">
            <text class="func-entry-icon">🚪</text>
            <text class="func-entry-label">查看退出</text>
          </view>
        </view>
      </view>

      <!-- 2. 业主/租户登记 -->
      <view class="menu-section">
        <view class="section-header">
          <text class="section-icon">🏠</text>
          <text class="section-title">业主/租户登记</text>
        </view>
        <view class="func-entries">
          <template v-if="communityStore.hasCommunities">
            <view class="func-entry" hover-class="func-entry--hover" @click="startOwnerAuth">
              <text class="func-entry-icon">🏠</text>
              <text class="func-entry-label">业主登记</text>
            </view>
            <view class="func-entry" hover-class="func-entry--hover" @click="startTenantAuth">
              <text class="func-entry-icon">🏠</text>
              <text class="func-entry-label">租户登记</text>
            </view>
          </template>
          <view v-else class="func-entry func-entry--disabled">
            <text class="func-entry-icon">🏠</text>
            <text class="func-entry-label">请先加入小区</text>
          </view>
        </view>
      </view>

      <!-- 3. 新增身份 -->
      <view class="menu-section">
        <view class="section-header">
          <text class="section-icon">🪪</text>
          <text class="section-title">新增身份</text>
        </view>
        <view class="func-entries">
          <view class="func-entry" hover-class="func-entry--hover" @click="onOwnerAuth">
            <text class="func-entry-icon">🪪</text>
            <text class="func-entry-label">业主认证</text>
          </view>
          <view class="func-entry" hover-class="func-entry--hover" @click="applyForRole('grid_worker')">
            <text class="func-entry-icon">🪪</text>
            <text class="func-entry-label">网格员认证</text>
          </view>
          <view class="func-entry" hover-class="func-entry--hover" @click="applyForRole('property_admin')">
            <text class="func-entry-icon">🪪</text>
            <text class="func-entry-label">物业管理员认证</text>
          </view>
          <view class="func-entry" hover-class="func-entry--hover" @click="applyForRole('community_admin')">
            <text class="func-entry-icon">🪪</text>
            <text class="func-entry-label">社区管理员认证</text>
          </view>
          <view class="func-entry" hover-class="func-entry--hover" @click="applyForRole('merchant')">
            <text class="func-entry-icon">🪪</text>
            <text class="func-entry-label">商家认证</text>
          </view>
        </view>
      </view>

      <!-- 4. 账号管理 -->
      <view class="menu-section">
        <view class="section-header">
          <text class="section-icon">⚙️</text>
          <text class="section-title">账号管理</text>
        </view>
        <view class="func-entries">
          <view class="func-entry" hover-class="func-entry--hover" @click="onLogout">
            <text class="func-entry-icon">🚪</text>
            <text class="func-entry-label">退出登录</text>
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

          <text class="modal-label" style="margin-top: 0.75rem;">登记房号</text>
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
import { getUserProfile, logout } from '@/api/identity';
import { applyRole, bindResidence, getUserMemberships, getUserRoles } from '@/api/user';
import { getDeviceId } from '@/utils/device';

const userStore = useUserStore();
const communityStore = useCommunityStore();

const pageLoading = ref(true);
// 用户已获取的角色列表（含认证状态 status：0未认证/1待审/2已认证/3驳回/4过期）
const userRoles = ref<any[]>([]);

// Identity / role state
const showBindResidence = ref(false);
const authTarget = ref<'owner' | 'tenant'>('owner');
const bindCommunityId = ref('');
const bindBuilding = ref('');
const bindUnit = ref('');
const bindRoom = ref('');

const hasOwnerRole = computed(() => {
  // 拥有已认证（status=2）的业主角色
  return userRoles.value.some(r => r.role_code === 'owner' && r.verf_status === 2);
});

// Phone display (full number from store, fallback to storage)
const displayPhone = computed(() => {
  const phone = userStore.user?.phone;
  if (phone) return phone;
  try {
    // user_phone 由 auth-flow.handleAuthSuccess 在登录/注册成功时写入（登录输入手机号兜底）
    return (uni.getStorageSync('user_phone') as string) || '未绑定手机号';
  } catch {
    return '未绑定手机号';
  }
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

// 业主认证：与网格员等角色一致，直接申请 owner 角色；
// 已有已认证业主角色（verf_status=2）则不重复申请（toast 提示）
function onOwnerAuth() {
  if (hasOwnerRole.value) {
    uni.showToast({ title: '已是业主', icon: 'none' });
    return;
  }
  applyForRole('owner');
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

// 退出登录：showModal 确认 → 调后端注销当前设备会话 → 清本地 token/user → 回到登录页（=退出页）
async function onLogout() {
  uni.showModal({
    title: '退出登录',
    content: '确定要退出登录吗？',
    success: async (res) => {
      if (!res.confirm) return;
      try {
        await logout(getDeviceId());
        userStore.logout();
        uni.reLaunch({ url: '/pages/login/login' });
      } catch (e) {
        // SEE: [[axios-network-error-raw-message-toast]] — toast 用固定中文文案，不取 e.message 原文
        console.error('[my] 退出登录失败', e);
        uni.showToast({ title: '退出登录失败', icon: 'none' });
      }
    },
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
    // 加载用户角色（用于判断是否已有 owner 角色，决定业主认证是否可重复申请）
    try {
      userRoles.value = await getUserRoles();
    } catch (e) {
      console.warn('[my] Failed to load user roles:', e);
    }
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
  padding: 6.25rem 0;
}

.loading-text {
  font-size: 0.875rem;
  color: $uni-text-color-placeholder;
}

// ---- Login prompt ----
.login-prompt {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 6.25rem 0;
}

.login-avatar {
  width: 4.375rem;
  height: 4.375rem;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 0.09375rem solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 0.125rem 0.625rem rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.login-avatar-emoji {
  font-size: 2.25rem;
}

.login-text {
  font-size: 1rem;
  color: #B8956A;
}

// ---- Gradient Header ----
.header {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 2.5rem 0 1.875rem;
  background: linear-gradient(160deg, #D4B896 0%, #E8DCCF 30%, #FAF8F5 55%, #FFFFFF 80%);
}

.avatar-circle {
  width: 4.375rem;
  height: 4.375rem;
  border-radius: 50%;
  background-color: rgba(255, 255, 255, 0.8);
  border: 0.09375rem solid rgba(184, 149, 106, 0.25);
  box-shadow: 0 0.125rem 0.625rem rgba(184, 149, 106, 0.12);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.625rem;
}

.avatar-emoji {
  font-size: 2.25rem;
}

.header-name {
  font-size: 0.75rem;
  color: #A6988A;
  margin-bottom: 0.25rem;
}

.header-phone {
  font-size: 1rem;
  font-weight: 600;
  color: #3D3226;
}

// ---- Sections (title + icon grid) ----
.menu-section {
  padding: 0 1.25rem;
  margin-top: 0.625rem;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-bottom: 0.5rem;
}

.section-icon {
  font-size: 1.125rem;
}

.section-title {
  font-size: 1.0625rem;
  font-weight: 600;
  color: #3D3226;
}

// 图标网格（参考首页 notice.vue .func-entries）：4 列，icon 1.375rem + label 0.75rem，卡片底
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
  background-color: #FAF8F5;
  border-radius: 0.5rem;
  padding: 0.75rem 0.25rem;
  box-shadow: $uni-shadow-sm;

  &--hover {
    background-color: rgba(184, 149, 106, 0.08);
    transition: background-color 0.15s ease;
  }

  &--disabled {
    opacity: 0.5;
  }
}

.func-entry-icon {
  font-size: 1.375rem;
  margin-bottom: 0.3125rem;
}

.func-entry-label {
  font-size: 0.75rem;
  color: #3D3226;
  font-weight: 500;
  text-align: center;
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
  width: 20rem;
  background: #fff;
  border-radius: 0.625rem;
  padding: 1.25rem 1rem;
}

.modal-title {
  font-size: 1.0625rem;
  font-weight: 700;
  color: #3D3226;
  display: block;
  text-align: center;
  margin-bottom: 0.25rem;
}

.modal-sub {
  font-size: 0.75rem;
  color: #A6988A;
  display: block;
  text-align: center;
  margin-bottom: 0.75rem;
}

.modal-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: #3D3226;
  display: block;
  margin-bottom: 0.5rem;
}

// ---- Community Picker ----
.community-picker {
  max-height: 7.5rem;
  margin-bottom: 0.25rem;
}

.community-option {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.5rem 0.625rem;
  background: #FAF8F5;
  border-radius: 0.3125rem;
  margin-bottom: 0.25rem;
  font-size: 0.8125rem;
  color: #3D3226;
  border: 0.0625rem solid transparent;

  &--selected {
    border-color: #B8956A;
    background: rgba(184, 149, 106, 0.06);
  }
}

.community-check {
  font-size: 0.875rem;
  color: #B8956A;
  font-weight: 700;
}

.community-single {
  padding: 0.5rem 0.625rem;
  background: #FAF8F5;
  border-radius: 0.3125rem;
  font-size: 0.8125rem;
  color: #3D3226;
  text-align: center;
  margin-bottom: 0.25rem;
}

.address-example {
  background: #FAF8F5;
  border-radius: 0.3125rem;
  padding: 0.5rem 0.625rem;
  margin-bottom: 0.75rem;
  text-align: center;
}

.address-example text {
  font-size: 0.75rem;
  color: #8C7B6B;
}

.address-inputs-row {
  display: flex;
  align-items: flex-start;
  justify-content: center;
  gap: 0.375rem;
  margin-bottom: 1rem;
}

.input-col {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.input-label {
  font-size: 0.6875rem;
  color: #8C7B6B;
  margin-bottom: 0.25rem;
}

.addr-input {
  width: 100%;
  height: 2.25rem;
  background: #FAF8F5;
  border: 0.0625rem solid #E8DCCF;
  border-radius: 0.375rem;
  text-align: center;
  font-size: 0.875rem;
  color: #3D3226;
  padding: 0 0.25rem;
}

.input-sep {
  font-size: 1rem;
  color: #CCC4BA;
  line-height: 2.25rem;
  padding-top: 0.9375rem;
}

.modal-btns {
  display: flex;
  gap: 0.5rem;
}

.btn-cancel {
  flex: 1;
  height: 2.5rem;
  border-radius: 1.25rem;
  background: #F0EBE3;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  color: #8C7B6B;
}

.btn-confirm {
  flex: 1;
  height: 2.5rem;
  border-radius: 1.25rem;
  background: linear-gradient(135deg, #B8956A, #D4B896);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  color: #fff;
  font-weight: 600;
}
</style>
