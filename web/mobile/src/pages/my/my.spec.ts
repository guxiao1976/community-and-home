// Component test — my.vue 退出登录流程 + 业主认证分支。
// 退出登录：showModal 确认 → await logout(getDeviceId()) → userStore.logout()（清 token/user）
//   → reLaunch('/pages/login/login')；取消分支不动作；logout 接口失败 → 仍本地登出（清 token+跳登录页）。
// 业主认证：已有业主角色（role_code=owner 且 verf_status=2）→ toast「已是业主」不重复申请；
//   否则 applyForRole('owner')（与网格员等一致，需 currentCommunityId）。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

vi.mock('@/api/identity', () => ({
  getUserProfile: vi.fn(),
  logout: vi.fn(),
}));

vi.mock('@/api/user', () => ({
  applyRole: vi.fn(),
  bindResidence: vi.fn(),
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getUserRoles: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  setCurrentCommunity: vi.fn(),
  getAppState: vi.fn().mockResolvedValue({ current_community_id: '0', updated_at: 0 }),
}));

vi.mock('@common/utils/auth', () => ({
  getAccessToken: vi.fn(() => null),
  getRefreshToken: vi.fn(() => null),
  setTokens: vi.fn(),
  clearTokens: vi.fn(),
  isAuthenticated: vi.fn(() => true),
}));

vi.mock('@/utils/device', () => ({
  getDeviceId: vi.fn(() => 'dev-1'),
}));

import { logout } from '@/api/identity';
import { applyRole, getUserRoles } from '@/api/user';
import { clearTokens } from '@common/utils/auth';
import { useUserStore } from '@/stores/user';
import { useCommunityStore } from '@/stores/community';
import MyPage from './my.vue';

let pinia: Pinia;

// uni 全局 stub 未内置 showModal（vitest.setup.ts），此处按测试需要注入
function mockShowModal(confirm: boolean) {
  (uni as any).showModal = vi.fn((opts: any) => {
    opts.success?.({ confirm });
  });
}

async function mountPage(roles: any[] = []): Promise<VueWrapper> {
  const store = useUserStore();
  store.setUser({ id: 'u1', nickname: '测试用户', phone: '13800001111' } as any);
  (getUserRoles as any).mockResolvedValue(roles);
  const wrapper = mount(MyPage, { global: { plugins: [pinia] } });
  await flushPromises();
  return wrapper;
}

describe('my page — 退出登录', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (logout as any).mockResolvedValue({});
    uni.removeStorageSync('current_community_id');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('确认退出 → 调 logout(getDeviceId()) + 清 token + 清 user_phone + reLaunch 登录页', async () => {
    mockShowModal(true);
    uni.setStorageSync('user_phone', '13800138000'); // 登录期写入的兜底缓存（auth-flow）
    const wrapper = await mountPage();
    const store = useUserStore();
    store.accessToken = 'at';

    await (wrapper.vm as any).onLogout();
    await flushPromises();

    expect(logout).toHaveBeenCalledWith('dev-1');
    expect(clearTokens).toHaveBeenCalled();
    expect(store.accessToken).toBe(null);
    // 退出必须清登录期兜底缓存，防共享设备跨账号串号（SEE: [[logout-clear-login-cache]]）
    expect(uni.getStorageSync('user_phone')).toBe('');
    expect(uni.reLaunch).toHaveBeenCalledWith({ url: '/pages/login/login' });
  });

  it('取消退出 → 不调 logout、不清 token、不跳转', async () => {
    mockShowModal(false);
    const wrapper = await mountPage();
    const store = useUserStore();
    store.accessToken = 'at';

    await (wrapper.vm as any).onLogout();
    await flushPromises();

    expect(logout).not.toHaveBeenCalled();
    expect(clearTokens).not.toHaveBeenCalled();
    expect(store.accessToken).toBe('at');
    expect(uni.reLaunch).not.toHaveBeenCalled();
  });

  it('logout 接口失败 → 仍本地登出（清 token + 跳登录页）', async () => {
    // 后端注销失败（如 access token 过期 401 / refresh 失效 / 网络错）不阻塞本地登出：
    // 用户点退出就是要退出，本地清 token 并导航登录页仍正确（token 已失效时本地清理即足够）。
    mockShowModal(true);
    (logout as any).mockRejectedValue(new Error('logout 接口异常'));
    const wrapper = await mountPage();
    const store = useUserStore();
    store.accessToken = 'at';

    await (wrapper.vm as any).onLogout();
    await flushPromises();

    expect(store.accessToken).toBe(null);
    expect(uni.reLaunch).toHaveBeenCalledWith({ url: '/pages/login/login' });
  });
});

describe('my page — 业主认证分支', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (applyRole as any).mockResolvedValue({});
    uni.removeStorageSync('current_community_id');
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('未认证业主 → applyRole({ community_id, role_code: owner })', async () => {
    const wrapper = await mountPage([]);
    const communityStore = useCommunityStore();
    communityStore.addCommunity({ communityId: 'c1', communityName: 'A 小区' });

    await (wrapper.vm as any).onOwnerAuth();

    expect(applyRole).toHaveBeenCalledWith({ community_id: 'c1', role_code: 'owner' });
  });

  it('已有业主角色（verf_status=2）→ toast「已是业主」，不重复申请', async () => {
    const wrapper = await mountPage([{ role_code: 'owner', verf_status: 2 }]);
    const communityStore = useCommunityStore();
    communityStore.addCommunity({ communityId: 'c1', communityName: 'A 小区' });

    await (wrapper.vm as any).onOwnerAuth();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '已是业主' }),
    );
    expect(applyRole).not.toHaveBeenCalled();
  });
});
