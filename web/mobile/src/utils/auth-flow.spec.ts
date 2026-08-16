// Unit tests for the shared post-auth flow (src/utils/auth-flow.ts).
// onAuthSuccess 抽成共享函数，login.vue 与 agreement.vue 复用。
// 登录态修复：profile 拉取失败 → console.error + toast 明确提示（不再静默导致 user=null），
// 但仍继续后续跳转（token 已存，App.vue onLaunch / 页面懒加载会再恢复 profile）。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §2
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  getUserMemberships: vi.fn(),
}));

vi.mock('@/api/identity', () => ({
  getUserProfile: vi.fn(),
}));

vi.mock('@common/utils/auth', () => ({
  getAccessToken: vi.fn(() => null),
  getRefreshToken: vi.fn(() => null),
  setTokens: vi.fn(),
  clearTokens: vi.fn(),
  isAuthenticated: vi.fn(() => false),
}));

import { getUserProfile } from '@/api/identity';
import { getUserMemberships } from '@/api/user';
import { handleAuthSuccess } from './auth-flow';
import { useUserStore } from '@/stores/user';

let pinia: Pinia;

describe('auth-flow — handleAuthSuccess', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('保存 token + 拉取 profile 并 setUser，无小区 → redirectTo 加入小区', async () => {
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockResolvedValue([]);
    const store = useUserStore();

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(store.accessToken).toBe('at');
    expect(store.user?.nickname).toBe('测试用户');
    expect(uni.redirectTo).toHaveBeenCalledWith({ url: '/pages/join-community/join-community' });
  });

  it('profile 拉取失败 → 单条合并 toast（icon:none），不再弹纯净「登录成功」success toast（REQ-TOAST-1）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getUserProfile as any).mockRejectedValue(new Error('profile 接口异常'));
    (getUserMemberships as any).mockResolvedValue([{ id: 'm1', community_id: 'c1' }]);
    const store = useUserStore();

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(errSpy).toHaveBeenCalled();
    // 合并提示同时表达成功与失败，icon:none（非 success 打勾，避免长文案截断）
    expect(uni.showToast).toHaveBeenCalledTimes(1);
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '登录成功，但资料加载失败', icon: 'none' }),
    );
    // 失败分支绝不出现纯净「登录成功」success toast 覆盖
    expect(uni.showToast).not.toHaveBeenCalledWith(
      expect.objectContaining({ title: '登录成功', icon: 'success' }),
    );
    expect(store.user).toBe(null);
    // token 已存、页面懒加载/App.vue 会恢复 profile，因此继续跳转不阻塞
    expect(uni.switchTab).toHaveBeenCalledWith({ url: '/pages/notice/notice' });
  });

  it('profile 拉取成功 → 纯净「登录成功」success toast（REQ-TOAST-1 成功路径）', async () => {
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockResolvedValue([{ id: 'm1', community_id: 'c1' }]);

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(uni.showToast).toHaveBeenCalledTimes(1);
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '登录成功', icon: 'success' }),
    );
    // 成功路径不出现合并/失败 toast
    expect(uni.showToast).not.toHaveBeenCalledWith(
      expect.objectContaining({ title: '登录成功，但资料加载失败' }),
    );
  });

  it('有小区 → switchTab 通知页', async () => {
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockResolvedValue([{ id: 'm1', community_id: 'c1' }]);

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(uni.switchTab).toHaveBeenCalledWith({ url: '/pages/notice/notice' });
  });

  it('getUserMemberships 失败 → console.warn + 默认加入小区（不静默吞错）', async () => {
    // SEE: [[verify-api-before-calling]] — 禁止空 catch 静默吞错：至少打日志，避免已有小区用户被静默误导
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockRejectedValue(new Error('membership 接口异常'));
    const warnSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
    const store = useUserStore();

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(warnSpy).toHaveBeenCalledWith(
      expect.stringContaining('[auth-flow] 小区检查失败'),
      expect.any(Error),
    );
    // membership 检查失败无法确认已有小区 → 默认走加入小区流程
    expect(uni.redirectTo).toHaveBeenCalledWith({ url: '/pages/join-community/join-community' });
    warnSpy.mockRestore();
  });

  it('opts.phone 提供 → 登录成功后写入 uni storage user_phone（手机号兜底）', async () => {
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockResolvedValue([]);
    const setStorageSpy = vi.spyOn(uni, 'setStorageSync');

    const p = handleAuthSuccess(
      { accessToken: 'at', refreshToken: 'rt', expiresAt: 123 },
      { phone: '13800001111' },
    );
    await vi.advanceTimersByTimeAsync(800);
    await p;

    // my.vue displayPhone 的 storage 兜底来源（后端 profile 对本人也可能脱敏，前端以登录输入手机号兜底）
    expect(setStorageSpy).toHaveBeenCalledWith('user_phone', '13800001111');
    setStorageSpy.mockRestore();
  });

  it('opts.phone 未提供 → 不写入 user_phone（避免误写空值）', async () => {
    (getUserProfile as any).mockResolvedValue({ id: 'u1', nickname: '测试用户' });
    (getUserMemberships as any).mockResolvedValue([]);
    const setStorageSpy = vi.spyOn(uni, 'setStorageSync');

    const p = handleAuthSuccess({ accessToken: 'at', refreshToken: 'rt', expiresAt: 123 });
    await vi.advanceTimersByTimeAsync(800);
    await p;

    expect(setStorageSpy).not.toHaveBeenCalledWith('user_phone', expect.any(String));
    setStorageSpy.mockRestore();
  });
});
