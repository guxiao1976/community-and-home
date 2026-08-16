// Unit tests for App.vue's restoreUserProfile (onLaunch 登录态恢复).
// 登录态修复：启动时若已登录（token 存在）但 user 未加载（登录后 profile 拉取失败 /
// tab 在登录前已挂载），全局恢复 profile，避免 isLoggedIn 误判未登录。
// 覆盖 4 分支：未登录不调 / user 未加载则 getUserProfile+setUser / user 已加载跳过 /
// getUserProfile 失败 console.error 不抛错。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setActivePinia, createPinia, type Pinia } from 'pinia';
import type { User } from '@common/types/identity';

vi.mock('@dcloudio/uni-app', () => ({
  onLaunch: vi.fn(),
  onShow: vi.fn(),
  onHide: vi.fn(),
}));

vi.mock('@common/utils/auth', () => ({
  getAccessToken: vi.fn(() => null),
  getRefreshToken: vi.fn(() => null),
  setTokens: vi.fn(),
  clearTokens: vi.fn(),
  isAuthenticated: vi.fn(),
}));

vi.mock('@/api/identity', () => ({
  getUserProfile: vi.fn(),
}));

import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';
import { useUserStore } from '@/stores/user';
import { restoreUserProfile } from './App.vue';

let pinia: Pinia;

function fakeUser(overrides: Partial<User> = {}): User {
  return {
    id: 'u1',
    phone: '13800000000',
    nickname: '测试用户',
    avatar: '',
    userType: 1,
    status: 1,
    verificationStatus: 0,
    scope: '',
    last_login_at: '',
    created_at: '',
    updated_at: '',
    deleted_at: 0,
    role_names: [],
    ...overrides,
  } as User;
}

describe('App.vue restoreUserProfile（onLaunch 登录态恢复）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  it('未登录（token 不存在）→ 不调 getUserProfile', async () => {
    (isAuthenticated as any).mockReturnValue(false);

    await restoreUserProfile();

    expect(getUserProfile).not.toHaveBeenCalled();
  });

  it('已登录但 user 未加载 → getUserProfile + setUser', async () => {
    (isAuthenticated as any).mockReturnValue(true);
    const profile = fakeUser({ nickname: '加载后昵称' });
    (getUserProfile as any).mockResolvedValue(profile);
    const store = useUserStore();

    await restoreUserProfile();

    expect(getUserProfile).toHaveBeenCalledTimes(1);
    expect(store.user).toEqual(profile);
  });

  it('已登录且 user 已加载 → 跳过 getUserProfile', async () => {
    (isAuthenticated as any).mockReturnValue(true);
    const store = useUserStore();
    store.setUser(fakeUser());

    await restoreUserProfile();

    expect(getUserProfile).not.toHaveBeenCalled();
  });

  it('getUserProfile 失败 → console.error 且不抛错', async () => {
    (isAuthenticated as any).mockReturnValue(true);
    (getUserProfile as any).mockRejectedValue(new Error('profile 获取失败'));
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    await expect(restoreUserProfile()).resolves.toBeUndefined();

    expect(errSpy).toHaveBeenCalled();
  });
});
