// Unit tests for the user store's isLoggedIn.
// 登录态修复：isLoggedIn 以 token 为权威（isAuthenticated），user 仅是 profile 缓存。
// 修复前 isLoggedIn = isAuthenticated() && user !== null —— 登录后 profile 拉取失败或
// tab 在登录前已挂载 → user=null → 页面误判未登录。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §1
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setActivePinia, createPinia, type Pinia } from 'pinia';
import type { User } from '@common/types/identity';

vi.mock('@common/utils/auth', () => ({
  getAccessToken: vi.fn(),
  getRefreshToken: vi.fn(),
  setTokens: vi.fn(),
  clearTokens: vi.fn(),
  isAuthenticated: vi.fn(),
}));

import { isAuthenticated } from '@common/utils/auth';
import { useUserStore } from './user';

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

describe('user store — isLoggedIn（token 权威）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  it('已登录（token 存在）但 user 未加载（user=null）→ isLoggedIn=true', () => {
    (isAuthenticated as any).mockReturnValue(true);
    const store = useUserStore();
    expect(store.user).toBe(null);
    expect(store.isLoggedIn).toBe(true);
  });

  it('未登录（无 token）即使 user 缓存存在 → isLoggedIn=false', () => {
    (isAuthenticated as any).mockReturnValue(false);
    const store = useUserStore();
    store.setUser(fakeUser());
    expect(store.isLoggedIn).toBe(false);
  });

  it('已登录且有 user → isLoggedIn=true', () => {
    (isAuthenticated as any).mockReturnValue(true);
    const store = useUserStore();
    store.setUser(fakeUser());
    expect(store.isLoggedIn).toBe(true);
  });
});
