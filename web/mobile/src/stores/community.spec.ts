// Unit tests for the community store's switchCommunity action.
// switchCommunity is a real logic function: it persists the new current community
// to the backend (setCurrentCommunity) and only mutates local state on success;
// on backend rejection (e.g. 10015 out-of-scope) it rethrows and leaves
// currentCommunityId unchanged.
// SEE: [[frontend-business-rule-hardcode]] — 切换校验权威在后端，前端仅消费 10015
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §1
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  getUserMemberships: vi.fn().mockResolvedValue([]),
  setCurrentCommunity: vi.fn(),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  getAppState: vi.fn().mockResolvedValue({ current_community_id: '0', updated_at: 0 }),
}));

import { setCurrentCommunity, getUserMemberships, getAppState, getResidentialAreasByIds } from '@/api/user';
import { useCommunityStore } from './community';

let pinia: Pinia;

describe('community store — switchCommunity', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (setCurrentCommunity as any).mockResolvedValue(undefined);
    uni.setStorageSync('current_community_id', '');
  });

  it('persists to backend then updates local state + storage on success', async () => {
    const store = useCommunityStore();
    store.addCommunity({ communityId: 'c1', communityName: 'A 小区' });
    store.addCommunity({ communityId: 'c2', communityName: 'B 小区' });
    store.currentCommunityId = 'c1';

    await store.switchCommunity('c2');

    expect(setCurrentCommunity).toHaveBeenCalledWith('c2');
    expect(store.currentCommunityId).toBe('c2');
    expect(uni.getStorageSync('current_community_id')).toBe('c2');
  });

  it('keeps currentCommunityId unchanged when backend rejects (e.g. 10015 out of scope)', async () => {
    const store = useCommunityStore();
    store.addCommunity({ communityId: 'c1', communityName: 'A 小区' });
    store.addCommunity({ communityId: 'c2', communityName: 'B 小区' });
    store.currentCommunityId = 'c1';
    uni.setStorageSync('current_community_id', 'c1');
    (setCurrentCommunity as any).mockRejectedValue(
      Object.assign(new Error('目标小区不在数据范围'), { code: 10015 }),
    );

    await expect(store.switchCommunity('c2')).rejects.toMatchObject({ code: 10015 });

    // 本地 currentCommunityId 保持不变，且未落 storage
    expect(store.currentCommunityId).toBe('c1');
    expect(uni.getStorageSync('current_community_id')).toBe('c1');
  });

  it('no-ops (no backend call) when switching to a non-member community', async () => {
    const store = useCommunityStore();
    store.addCommunity({ communityId: 'c1', communityName: 'A 小区' });

    await store.switchCommunity('c99');

    expect(setCurrentCommunity).not.toHaveBeenCalled();
    expect(store.currentCommunityId).toBe('c1');
  });
});

describe('community store — loadMemberships 服务端权威（getAppState）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    uni.setStorageSync('current_community_id', '');
    (getUserMemberships as any).mockResolvedValue([
      { id: 'm1', community_id: 'c1' },
      { id: 'm2', community_id: 'c2' },
    ]);
    (getResidentialAreasByIds as any).mockResolvedValue([
      { id: 'c1', name: 'A 小区' },
      { id: 'c2', name: 'B 小区' },
    ]);
  });

  it('后端 current_community_id 存在于 memberships → 采用并保存（跨设备一致，修复本地 storage 陈旧）', async () => {
    // 本地 storage 陈旧指向 c1，后端权威为 c2
    uni.setStorageSync('current_community_id', 'c1');
    (getAppState as any).mockResolvedValue({ current_community_id: 'c2', updated_at: 0 });

    const store = useCommunityStore();
    await store.loadMemberships();

    expect(store.currentCommunityId).toBe('c2');
    expect(uni.getStorageSync('current_community_id')).toBe('c2');
  });

  it('getAppState 返回 0（未设置）→ 降级本地回退（无选中取第一个小区）', async () => {
    (getAppState as any).mockResolvedValue({ current_community_id: '0', updated_at: 0 });

    const store = useCommunityStore();
    await store.loadMemberships();

    expect(store.currentCommunityId).toBe('c1');
    expect(uni.getStorageSync('current_community_id')).toBe('c1');
  });

  it('getAppState 请求失败 → 容错降级忽略 + console.error 留痕（禁止静默吞错）', async () => {
    // SEE: [[verify-api-before-calling]] — 禁止空 catch 静默吞错：至少打日志，运维/QA 可感知 app-state 接口故障
    uni.setStorageSync('current_community_id', 'c1');
    (getAppState as any).mockRejectedValue(new Error('app-state 接口不可用'));
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    const store = useCommunityStore();
    await store.loadMemberships();

    expect(errSpy).toHaveBeenCalledWith(
      expect.stringContaining('[community] getAppState 获取失败，降级本地'),
      expect.any(Error),
    );
    expect(store.currentCommunityId).toBe('c1');
    expect(store.communities.length).toBe(2);
    errSpy.mockRestore();
  });

  it('后端 current_community_id 不在 memberships（已退出）→ 忽略服务端值，走本地回退', async () => {
    uni.setStorageSync('current_community_id', 'c1');
    (getAppState as any).mockResolvedValue({ current_community_id: 'c999', updated_at: 0 });

    const store = useCommunityStore();
    await store.loadMemberships();

    // c999 不在 memberships，不能采用 → 保留本地 c1
    expect(store.currentCommunityId).toBe('c1');
  });
});
