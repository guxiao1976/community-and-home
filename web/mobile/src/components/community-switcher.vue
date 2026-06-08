<template>
  <view class="cs-root">
    <!-- 渐变头部背景 -->
    <view class="cs-header">
      <!-- 切换器按钮 -->
      <view class="cs-trigger" @click.stop="toggle">
        <text class="cs-trigger-icon">🏘️</text>
        <view class="cs-trigger-body">
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
