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
}));

import { setCurrentCommunity } from '@/api/user';
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
