// 登录/注册成功后的共享处理流程（login.vue 与 agreement.vue 复用）。
// 登录态修复：profile 拉取失败不再静默继续导致 user=null——console.error + toast 明确提示，
// 但仍继续跳转（token 已存，App.vue onLaunch / 页面懒加载会再恢复 profile；isLoggedIn 以 token 为权威）。
import { getUserProfile } from '@/api/identity';
import { getUserMemberships } from '@/api/user';
import { useUserStore } from '@/stores/user';

export interface AuthSuccessResult {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

/**
 * 保存 token + 拉取 profile + 判断小区后跳转（注册完成自动登录一次）。
 * @param opts.onCompleted 跳转完成后回调（如复位 submitting）
 * @param opts.phone 用户登录/注册时输入的手机号：写入 uni storage 'user_phone'，
 *   my.vue displayPhone 以此兜底（后端 profile 对本人也可能脱敏，前端不依赖后端修复时序）。
 */
export async function handleAuthSuccess(
  loginRes: AuthSuccessResult,
  opts: { onCompleted?: () => void; phone?: string } = {},
): Promise<void> {
  const userStore = useUserStore();

  // 1. 保存 tokens
  userStore.setAuth(loginRes);

  // 1.1 手机号兜底：登录/注册成功后立即写入 storage（置于 profile 拉取前，
  // profile 失败也不丢失手机号）。仅当调用方明确传入 phone 时写入，避免误写空值。
  // SEE: [[verify-api-before-calling]] — 前端不依赖后端 profile 脱敏修复时序
  if (opts.phone) {
    uni.setStorageSync('user_phone', opts.phone);
  }

  // 2. 拉取 user profile
  let profileFailed = false;
  try {
    uni.showLoading({ title: '登录中...', mask: true });
    const user = await getUserProfile();
    userStore.setUser(user);
  } catch (err) {
    // 不静默：记录留痕（避免 user=null 导致页面误判未登录）；token 已存，
    // App.vue onLaunch / 页面懒加载会再恢复 profile。提示与成功合并为单条 toast（REQ-TOAST-1）。
    profileFailed = true;
    console.error('[auth-flow] 获取用户资料失败', err);
  } finally {
    uni.hideLoading();
  }

  // 3. 判断用户是否已加入小区，据此跳转
  let hasCommunities = false;
  try {
    const memberships = await getUserMemberships();
    hasCommunities = memberships.length > 0;
  } catch (e) {
    // membership 检查失败默认走加入小区；必须留痕，否则接口故障时已有小区用户会被静默误导。
    // SEE: [[verify-api-before-calling]] — 禁止空 catch 静默吞错
    console.warn('[auth-flow] 小区检查失败，默认加入小区', e);
  }

  // profile 拉取失败 → 单条合并 toast（icon:none）：同时表达登录成功与资料加载失败，
  // 不再先弹失败提示再被后续「登录成功」(icon:success) 覆盖；文案不承诺自动恢复
  // （profile 恢复仅发生于 App.vue onLaunch 启动时与 mine 页面懒加载，登录流程本身不承诺）。
  // icon:none 文字型 toast，非 success 打勾：不误导「完全成功」且长文案不截断。
  // SEE: [[axios-network-error-raw-message-toast]] — toast 一律用固定中文文案，不取 e.message 原文
  if (profileFailed) {
    uni.showToast({ title: '登录成功，但资料加载失败', icon: 'none', duration: 1500 });
  } else {
    uni.showToast({ title: '登录成功', icon: 'success', duration: 1500 });
  }
  setTimeout(() => {
    if (hasCommunities) {
      uni.switchTab({ url: '/pages/notice/notice' });
    } else {
      uni.redirectTo({ url: '/pages/join-community/join-community' });
    }
    opts.onCompleted?.();
  }, 800);
}
