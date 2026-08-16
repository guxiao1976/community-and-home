<template>
  <view class="page">
    <!-- 加载中 -->
    <view v-if="loading" class="status-wrap">
      <text class="status-text">加载中...</text>
    </view>

    <!-- 加载失败（REQ-CLP-1 场景 4，禁止静默吞错） -->
    <view v-else-if="loadError" class="status-wrap">
      <text class="status-text">加载失败，请稍后重试</text>
    </view>

    <!-- 空态（REQ-CLP-1 场景 3，无种子数据） -->
    <view v-else-if="contacts.length === 0" class="status-wrap">
      <text class="empty-icon">📞</text>
      <text class="status-text">暂无联络信息</text>
    </view>

    <!-- 拨号网格（类别图标 + 类别名 + 电话，样式沿用首页原联络网格） -->
    <view v-else class="contact-grid">
      <view
        v-for="contact in contacts"
        :key="contact.id"
        class="contact-cell"
        @click="onCall(contact)"
      >
        <text class="contact-icon">{{ getContactCategoryIcon(contact.category) }}</text>
        <view class="contact-body">
          <text class="contact-name">{{ getContactCategoryName(contact.category) || contact.name }}</text>
          <text class="contact-phone">{{ contact.phone }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { getContacts, getContactCategoryName, getContactCategoryIcon } from '@/api/community';
import type { Contact } from '@/api/community';
import { useCommunityStore } from '@/stores/community';

const communityStore = useCommunityStore();

const contacts = ref<Contact[]>([]);
const loading = ref(true);
const loadError = ref(false);

function onCall(contact: Contact) {
  if (contact.phone) {
    uni.makePhoneCall({ phoneNumber: contact.phone });
  }
}

async function fetchContacts(communityId: string) {
  loading.value = true;
  loadError.value = false;
  try {
    // SEE: [[verify-api-before-calling]] — GET /api/community/contacts 已在 graph-context 确认
    const data = await getContacts(communityId);
    contacts.value = data.contacts || [];
  } catch (e) {
    console.error('[contact-list] 联络列表加载失败', e);
    contacts.value = [];
    loadError.value = true;
  } finally {
    loading.value = false;
  }
}

onLoad(() => {
  const cid = communityStore.currentCommunityId;
  if (cid) {
    fetchContacts(cid);
  } else {
    loading.value = false;
  }
});
</script>

<style scoped lang="scss">
.page {
  min-height: 100vh;
  background-color: $uni-bg-color;
  padding: 24rpx 32rpx;
}

// ---- 拨号网格 ----
.contact-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16rpx;
}

.contact-cell {
  display: flex;
  align-items: center;
  background-color: $uni-bg-color-card;
  border-radius: 12rpx;
  padding: 20rpx 18rpx;
  box-shadow: $uni-shadow-sm;
}

.contact-icon {
  font-size: 40rpx;
  margin-right: 14rpx;
  flex-shrink: 0;
}

.contact-body {
  flex: 1;
  min-width: 0;
}

.contact-name {
  display: block;
  font-size: 24rpx;
  color: $uni-text-color-grey;
  margin-bottom: 4rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.contact-phone {
  display: block;
  font-size: 30rpx;
  font-weight: 500;
  color: $uni-text-color;
}

// ---- Status ----
.status-wrap {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 0;
}

.empty-icon {
  font-size: 72rpx;
  margin-bottom: 16rpx;
  opacity: 0.6;
}

.status-text {
  font-size: 28rpx;
  color: $uni-text-color-placeholder;
}
</style>
