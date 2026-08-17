<template>
  <view class="page">
    <view class="header">
      <text class="header-title">加入 {{ pendingCommunity?.communityName || '小区' }}</text>
      <text v-if="pendingCommunity?.address" class="header-addr">{{ pendingCommunity.address }}</text>
    </view>

    <!-- 房屋权属：自有/租住（必填） -->
    <view class="form-card">
      <text class="form-label">房屋权属 <text class="required">*</text></text>
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

      <text class="form-label">楼号 <text class="required">*</text></text>
      <input class="join-form-input" v-model="joinForm.building" type="number" placeholder="如 3" />
      <text v-if="joinFormErrors.building" class="field-error">{{ joinFormErrors.building }}</text>

      <text class="form-label">单元号 <text class="required">*</text></text>
      <input class="join-form-input" v-model="joinForm.unit" type="number" placeholder="如 1" />
      <text v-if="joinFormErrors.unit" class="field-error">{{ joinFormErrors.unit }}</text>

      <text class="form-label">房号 <text class="required">*</text></text>
      <input class="join-form-input" v-model="joinForm.room" type="number" placeholder="如 502" />
      <text v-if="joinFormErrors.room" class="field-error">{{ joinFormErrors.room }}</text>

      <button class="btn confirm-join-btn" :disabled="submitting" @click="confirmJoin">确认加入</button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { joinCommunity } from '@/api/user';
import { useCommunityStore } from '@/stores/community';
import { readPendingJoin, clearPendingJoin, type PendingJoin } from '@/utils/pending-join';
import {
  OWNERSHIP_OPTIONS, validateJoinForm, joinFormToPayload,
  type JoinFormErrors, type JoinFormState,
} from '../join-community/join-form';

const communityStore = useCommunityStore();

const pendingCommunity = ref<PendingJoin | null>(null);
const joinForm = ref<JoinFormState>({ building: '', unit: '', room: '', ownership: null });
const joinFormErrors = ref<JoinFormErrors>({});
const submitting = ref(false);

onMounted(() => {
  // 待加入小区来自 join-community → join-choice 的分流，pending-join 唯一契约源随行
  // SEE: [[frontend-cross-page-storage-contract]] — 跨页临时数据收敛到共享模块
  pendingCommunity.value = readPendingJoin();
});

async function confirmJoin() {
  const pending = pendingCommunity.value;
  if (!pending) {
    uni.showToast({ title: '请先选择小区', icon: 'none' });
    return;
  }
  if (submitting.value) return;

  // 前端 UX 即时校验（复用 join-form）；后端 JoinCommunity 仍权威校验（防绕过）
  // SEE: [[frontend-business-rule-hardcode]] — 前端仅展示层校验，权威在后端
  const result = validateJoinForm(joinForm.value);
  joinFormErrors.value = result.errors;
  if (!result.valid) return;

  submitting.value = true;
  try {
    uni.hideLoading(); // clear any lingering loading state
    uni.showLoading({ title: '加入中...', mask: true });
    const { building, unit, room, ownership } = joinFormToPayload(joinForm.value);
    // 后端 JoinCommunity 已按权属自动授权 owner（自有）/tenant（租住）角色（status=0 未认证），
    // 前端不得再自申请 applyRole('owner')——对租住用户会误授 owner（越权）。
    // SEE: [[join-auto-grant-vs-frontend-reapply-role-mismatch]]
    await joinCommunity(pending.communityId, building, unit, room, ownership);

    communityStore.addCommunity({
      communityId: pending.communityId,
      communityName: pending.communityName,
      address: pending.address,
    });
    clearPendingJoin();
    pendingCommunity.value = null;
    uni.showToast({ title: '加入成功', icon: 'success' });
    uni.switchTab({ url: '/pages/notice/notice' });
  } catch (e: any) {
    const msg = e?.message || e?.msg || '加入失败，请稍后重试';
    uni.showToast({ title: msg, icon: 'none', duration: 3000 });
  } finally {
    uni.hideLoading();
    submitting.value = false;
  }
}
</script>

<style scoped lang="scss">
.page { min-height: 100vh; background: #FFFFFF; padding: 0 2rem; }

.header { padding: 1.875rem 0 1.25rem; text-align: center;
  .header-title { display: block; font-size: 1.375rem; font-weight: 700; color: $uni-text-color; }
  .header-addr { display: block; font-size: 0.75rem; color: $uni-text-color-grey; margin-top: 0.25rem; }
}

.form-card { background: #FAF8F5; border-radius: 0.75rem; padding: 1.25rem 1rem; box-shadow: $uni-shadow-base; }
.form-label { display: block; font-size: 0.875rem; color: $uni-text-color; margin-bottom: 0.375rem;
  .required { color: $uni-color-error; }
}
.ownership-row { display: flex; gap: 0.625rem; margin-bottom: 0.25rem;
  .ownership-option { flex: 1; text-align: center; padding: 0.625rem 0; border-radius: 0.375rem; background: $uni-bg-color-input; color: $uni-text-color-grey; font-size: 0.875rem; border: 0.0625rem solid transparent;
    &.selected { background: rgba(184, 149, 106, 0.12); color: $uni-color-primary; border-color: $uni-color-primary; font-weight: 600; }
  }
}
.join-form-input { width: 100%; height: 2.5rem; background: $uni-bg-color-input; border-radius: 0.375rem; padding: 0 0.625rem; font-size: 0.875rem; color: $uni-text-color; box-sizing: border-box; margin-bottom: 0.25rem; }
.field-error { display: block; font-size: 0.6875rem; color: $uni-color-error; margin-bottom: 0.5rem; }

.btn { width: 100%; height: 2.75rem; border-radius: 1.375rem; background: linear-gradient(135deg, #B8956A, #D4B896); color: #fff; font-size: 0.9375rem; font-weight: 600; border: none; margin-top: 1rem;
  &::after { border: none; }
  &[disabled] { background: #E8E0D5; color: #CCC4BA; }
}
</style>
