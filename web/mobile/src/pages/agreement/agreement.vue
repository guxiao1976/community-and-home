<template>
  <view class="agreement-page">
    <!-- 顶部提示语 -->
    <view class="hint-box">
      <text class="hint-icon">📢</text>
      <text class="hint-text">您尚未注册账号，请阅读使用协议，勾选同意后注册。注册通过后，您可以登记为业主，也可以认证为网格员等其他身份，解锁 App 功能。</text>
    </view>

    <!-- 协议正文（只读可滚动文本框：内部滚动，不占满视口） -->
    <scroll-view class="agreement-box" scroll-y>
      <text class="agreement-title">《社区家园使用协议》</text>
      <text class="agreement-body">
        欢迎使用社区家园。在使用本社区 App 前，请仔细阅读以下条款：{'\n'}
        {'\n'}
        一、服务内容{'\n'}
        本平台为社区居民提供公告通知、邻里互动、寻失互助、便民联络等社区生活服务。{'\n'}
        {'\n'}
        二、用户信息{'\n'}
        您承诺提供真实、准确、完整的手机号等注册信息，并对其使用与保管负责。{'\n'}
        {'\n'}
        三、使用规范{'\n'}
        请勿利用平台发布违法违规、侵权、骚扰或虚假信息；请勿破坏社区秩序或侵犯他人合法权益。{'\n'}
        {'\n'}
        四、隐私保护{'\n'}
        我们将依法保护您的个人信息，仅在提供服务的必要范围内收集、使用与保存。{'\n'}
        {'\n'}
        五、服务变更与终止{'\n'}
        我们可能根据运营需要调整或终止部分服务，届时将依法履行告知义务。{'\n'}
        {'\n'}
        六、其他{'\n'}
        本协议未尽事宜，请参照国家相关法律法规及平台公示规则执行。
      </text>
    </scroll-view>

    <!-- 确认注册（固定底部，不滚动即可见） -->
    <view class="agreement-footer">
      <view class="agreement-check" @click="toggleAgreed">
        <view class="checkbox" :class="{ 'checkbox--checked': agreed }">
          <text v-if="agreed" class="check-mark">✓</text>
        </view>
        <view class="agreement-text">
          已阅并同意<text class="agreement-link">《社区家园使用协议》</text>
        </view>
      </view>
      <button
        class="confirm-btn"
        :disabled="submitting"
        :loading="submitting"
        @click="confirmRegister"
      >
        {{ submitting ? '注册中...' : '确认注册' }}
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { onLoad } from '@dcloudio/uni-app';
import { register } from '@/api/identity';
import { handleAuthSuccess } from '@/utils/auth-flow';
// SEE: [[sms-code-persist-localstorage]] — 一次性验证码经共享模块走内存态 + sessionStorage，不落 localStorage
// SEE: [[frontend-cross-page-storage-contract]] — key/结构收敛到共享契约源，不再内联 magic string
// SEE: [[cross-page-sensitive-temp-data-storage]] — 跨页一次性敏感数据优先内存态载体
import { readRegPending, clearRegPending, type RegPending } from '@/utils/reg-pending';

const agreed = ref(false);
const submitting = ref(false);
const pending = ref<RegPending | null>(null);

function toggleAgreed() {
  agreed.value = !agreed.value;
}

async function confirmRegister() {
  if (!agreed.value) {
    uni.showToast({ title: '请先阅读并同意使用协议', icon: 'none' });
    return;
  }
  if (!pending.value) {
    uni.showToast({ title: '注册信息已失效，请重新登录', icon: 'none' });
    return;
  }
  if (submitting.value) return;

  submitting.value = true;
  try {
    uni.showLoading({ title: '注册中...', mask: true });
    const regRes = await register({
      phone: pending.value.phone,
      smsCode: pending.value.smsCode,
      nickname: pending.value.nickname,
      deviceId: pending.value.deviceId,
    });
    uni.hideLoading();
    clearRegPending(); // 注册成功清除临时数据
    // 注册完成自动登录一次（保存 token + profile + 小区跳转）。
    // phone 传入 handleAuthSuccess 写入 user_phone（my.vue displayPhone 兜底）。
    // submitting 由 handleAuthSuccess 的 onCompleted 在跳转完成后复位，防止跳转窗口期二次点击触发重复注册
    await handleAuthSuccess(regRes, {
      phone: pending.value.phone,
      onCompleted: () => {
        submitting.value = false;
      },
    });
  } catch (err: unknown) {
    uni.hideLoading();
    submitting.value = false;
    const e = err as Error & { code?: number };
    const msg = e?.message || '';
    console.error('[agreement] 注册失败', err);
    // ① 手机号已注册（user-service CreateUser 10002）：账号已存在，清临时数据回登录页直接登录
    if (e?.code === 10002) {
      clearRegPending();
      uni.showToast({ title: '该手机号已注册，请直接登录', icon: 'none' });
      setTimeout(() => uni.navigateBack({ delta: 1 }), 1000);
      return;
    }
    // ② 超时/网络中断：请求结果不确定，账号可能已创建且验证码可能已消费——保留数据可重试，
    //    重试若命中 10002 会自动转登录；提示明确避免用户误以为必然失败。
    //    SEE: [[axios-network-error-raw-message-toast]]
    if (/timeout|ETIMEDOUT|ECONNABORTED|network/i.test(msg)) {
      uni.showToast({ title: '注册超时，账号可能已创建；请重试或返回登录', icon: 'none', duration: 2500 });
      return;
    }
    // ③ 其他业务错误
    uni.showToast({ title: '注册失败，请重试', icon: 'none' });
    // 保留临时数据，可重试
  }
}

onLoad(() => {
  pending.value = readRegPending();
  if (!pending.value) {
    uni.showToast({ title: '注册信息已失效，请重新登录', icon: 'none' });
    setTimeout(() => {
      uni.navigateBack({ delta: 1 });
    }, 1000);
  }
});
</script>

<style scoped lang="scss">
.agreement-page {
  /* 固定视口内可用高度（扣除导航栏 --window-top），footer 始终在屏内可见、不随协议内容顶出 */
  height: calc(100vh - var(--window-top, 0px));
  background: $uni-bg-color;
  display: flex;
  flex-direction: column;
  padding: 1rem 1rem 0;
  box-sizing: border-box;
}

/* 顶部提示语 */
.hint-box {
  display: flex;
  align-items: flex-start;
  gap: 0.375rem;
  background: rgba($uni-color-primary, 0.08);
  border-radius: 0.5rem;
  padding: 0.625rem 0.75rem;
  margin-bottom: 0.75rem;
  flex-shrink: 0;
}

.hint-icon {
  font-size: 0.875rem;
  flex-shrink: 0;
}

.hint-text {
  font-size: 0.75rem;
  color: $uni-text-color;
  line-height: 1.6;
}

/* 只读可滚动文本框：内部滚动；min-height:0 + overflow:hidden 允许 flex 收缩，使滚动框不占满视口 */
.agreement-box {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  border: 0.0625rem solid $uni-border-color;
  border-radius: 0.5rem;
  background: $uni-bg-color-card;
  padding: 1rem;
  box-sizing: border-box;
}

.agreement-title {
  display: block;
  font-size: 1.0625rem;
  font-weight: 600;
  color: $uni-text-color;
  margin-bottom: 0.75rem;
}

.agreement-body {
  display: block;
  font-size: 0.8125rem;
  line-height: 1.8;
  color: $uni-text-color-grey;
  white-space: pre-line;
}

.agreement-footer {
  /* 页级已有左右 padding；此处仅上下，footer 固定屏内 */
  padding: 0.75rem 0 calc(0.75rem + env(safe-area-inset-bottom));
  background: $uni-bg-color;
  box-shadow: 0 -0.125rem 0.75rem rgba(0, 0, 0, 0.05);
  flex-shrink: 0;
}

.agreement-check {
  display: flex;
  align-items: center;
  margin-bottom: 0.75rem;

  .checkbox {
    width: 1.125rem;
    height: 1.125rem;
    border-radius: 0.1875rem;
    border: 0.0625rem solid #c8c9cc;
    margin-right: 0.375rem;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;

    &--checked {
      background: $uni-color-primary;
      border-color: $uni-color-primary;

      .check-mark {
        color: #fff;
        font-size: 0.75rem;
        font-weight: 700;
      }
    }
  }

  .agreement-text {
    font-size: 0.75rem;
    color: $uni-text-color-grey;

    .agreement-link {
      color: $uni-color-primary;
    }
  }
}

.confirm-btn {
  width: 100%;
  height: 3rem;
  border-radius: 1.5rem;
  background: linear-gradient(135deg, #B8956A, #D4B896);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  border: none;
  display: flex;
  align-items: center;
  justify-content: center;

  &::after {
    border: none;
  }
}
</style>
