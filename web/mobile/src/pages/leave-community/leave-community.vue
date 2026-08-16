<template>
  <view class="page">
    <!-- Confirm Modal -->
    <view v-if="showModal" class="modal-overlay" @click="closeModal">
      <view class="modal-card" @click.stop>
        <view class="modal-icon">⚠️</view>
        <text class="modal-title">确认退出？</text>
        <text class="modal-desc">
          退出「{{ leavingCommunity?.communityName }}」后，您将不再接收该小区的通知和公告。可重新申请加入。
        </text>
        <view class="modal-actions">
          <view class="modal-btn modal-btn--cancel" @click="closeModal">
            <text>取消</text>
          </view>
          <view class="modal-btn modal-btn--confirm" @click="confirmLeave">
            <text>确认退出</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Header -->
    <view class="header">
      <view class="back-row" @click="goBack">
        <text class="back-icon">←</text>
        <text class="back-text">返回</text>
      </view>
      <text class="header-title">退出小区</text>
    </view>

    <!-- Warning Banner -->
    <view class="warning-banner">
      <text class="warning-icon">⚠️</text>
      <text class="warning-text">退出后将不再接收该小区的通知和公告。可重新申请加入。</text>
    </view>

    <!-- Loading -->
    <view v-if="loading" class="status-text">
      <text>加载中...</text>
    </view>

    <!-- Empty State -->
    <view v-else-if="communityStore.communities.length === 0" class="empty-state">
      <text class="empty-icon">🏘️</text>
      <text class="empty-text">暂未加入任何小区</text>
      <text class="empty-sub">加入小区后，可在此处管理退出</text>
    </view>

    <!-- Community List -->
    <view v-else class="community-list">
      <text class="list-title">已加入的小区（{{ communityStore.communities.length }}个）</text>

      <view
        v-for="c in communityStore.communities"
        :key="c.communityId"
        class="community-card"
      >
        <view class="card-left">
          <view class="card-name-row">
            <text class="card-emoji">🏘️</text>
            <text class="card-name">{{ c.communityName }}</text>
            <view v-if="c.communityId === communityStore.currentCommunityId" class="current-tag">
              <text>当前</text>
            </view>
          </view>
          <text v-if="c.address" class="card-addr">{{ c.address }}</text>
        </view>
        <view class="card-right">
          <view class="leave-btn" @click="openModal(c)">
            <text>退出</text>
          </view>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { leaveCommunity } from '@/api/user';
import { useCommunityStore, type CommunityInfo } from '@/stores/community';

const communityStore = useCommunityStore();
const loading = ref(false);

const showModal = ref(false);
const leavingCommunity = ref<CommunityInfo | null>(null);

onMounted(async () => {
  loading.value = true;
  try {
    await communityStore.loadMemberships();
  } catch (_) {
    // ignore
  } finally {
    loading.value = false;
  }
});

function goBack() {
  uni.navigateBack();
}

function openModal(community: CommunityInfo) {
  leavingCommunity.value = community;
  showModal.value = true;
}

function closeModal() {
  showModal.value = false;
  leavingCommunity.value = null;
}

async function confirmLeave() {
  if (!leavingCommunity.value) return;
  const id = leavingCommunity.value.communityId;

  try {
    uni.showLoading({ title: '退出中...', mask: true });
    await leaveCommunity(id);
    communityStore.removeCommunity(id);
    uni.hideLoading();
    uni.showToast({ title: '已退出小区', icon: 'success', duration: 1500 });
  } catch (_) {
    uni.hideLoading();
    uni.showToast({ title: '退出失败，请稍后重试', icon: 'none', duration: 2000 });
  } finally {
    closeModal();
  }
}
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background: #FAF8F5;
  padding: 0 1rem;
}

// Header
.header {
  padding-top: 0.625rem;
}

.back-row {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  margin-bottom: 0.5rem;

  .back-icon {
    font-size: 1rem;
    color: $uni-color-primary;
  }

  .back-text {
    font-size: 0.8125rem;
    color: $uni-text-color-grey;
  }
}

.header-title {
  display: block;
  font-size: 1.25rem;
  font-weight: 700;
  color: $uni-text-color;
  margin-bottom: 1.25rem;
}

// Warning Banner
.warning-banner {
  display: flex;
  align-items: flex-start;
  gap: 0.375rem;
  background: #FFF8F0;
  border: 0.03125rem solid #F0DCC0;
  border-radius: 0.375rem;
  padding: 0.75rem;
  margin-bottom: 1rem;

  .warning-icon {
    font-size: 1rem;
    flex-shrink: 0;
    line-height: 1.25rem;
  }

  .warning-text {
    font-size: 0.75rem;
    color: #8B7355;
    line-height: 1.1875rem;
  }
}

// Status
.status-text {
  text-align: center;
  padding: 3.75rem 0;
  font-size: 0.8125rem;
  color: $uni-text-color-placeholder;
}

// Empty State
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3.75rem 0;
  text-align: center;

  .empty-icon {
    font-size: 2.5rem;
    margin-bottom: 0.75rem;
  }

  .empty-text {
    font-size: 0.9375rem;
    font-weight: 600;
    color: $uni-text-color;
    margin-bottom: 0.375rem;
  }

  .empty-sub {
    font-size: 0.75rem;
    color: $uni-text-color-grey;
  }
}

// Community List
.community-list {
  .list-title {
    display: block;
    font-size: 0.8125rem;
    color: $uni-text-color-grey;
    margin-bottom: 0.625rem;
  }
}

.community-card {
  display: flex;
  align-items: center;
  background: #FFFFFF;
  border-radius: 0.5rem;
  padding: 0.875rem 0.75rem;
  margin-bottom: 0.5rem;
  box-shadow: $uni-shadow-sm;

  .card-left {
    flex: 1;
    overflow: hidden;
  }

  .card-name-row {
    display: flex;
    align-items: center;
    gap: 0.3125rem;
    margin-bottom: 0.25rem;
  }

  .card-emoji {
    font-size: 1.125rem;
  }

  .card-name {
    font-size: 0.9375rem;
    font-weight: 600;
    color: $uni-text-color;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .current-tag {
    flex-shrink: 0;
    padding: 0.125rem 0.4375rem;
    border-radius: 0.25rem;
    background: rgba(180, 150, 106, 0.12);

    text {
      font-size: 0.625rem;
      color: $uni-color-primary;
      font-weight: 500;
    }
  }

  .card-addr {
    font-size: 0.75rem;
    color: $uni-text-color-grey;
    display: block;
    margin-left: 1.4375rem;
  }

  .card-right {
    flex-shrink: 0;
    margin-left: 0.5rem;
  }

  .leave-btn {
    padding: 0.375rem 1rem;
    border-radius: 1rem;
    border: 0.0625rem solid #D4958A;
    background: #FFFFFF;

    text {
      font-size: 0.75rem;
      color: #D4958A;
      font-weight: 500;
    }
  }
}

// Modal
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.modal-card {
  width: 17.5rem;
  background: #FFFFFF;
  border-radius: 0.625rem;
  padding: 1.5rem 1.25rem 1.125rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  box-shadow: 0 0.5rem 2rem rgba(0, 0, 0, 0.12);
}

.modal-icon {
  font-size: 1.75rem;
  margin-bottom: 0.625rem;
}

.modal-title {
  font-size: 1.0625rem;
  font-weight: 700;
  color: $uni-text-color;
  margin-bottom: 0.5rem;
}

.modal-desc {
  font-size: 0.8125rem;
  color: $uni-text-color-grey;
  line-height: 1.25rem;
  margin-bottom: 1.125rem;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  width: 100%;
}

.modal-btn {
  flex: 1;
  height: 2.5rem;
  border-radius: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;

  text {
    font-size: 0.875rem;
    font-weight: 600;
  }

  &--cancel {
    background: #F5F0EA;

    text {
      color: $uni-text-color-grey;
    }
  }

  &--confirm {
    background: linear-gradient(135deg, #D4958A, #C57F74);

    text {
      color: #FFFFFF;
    }
  }
}
</style>
