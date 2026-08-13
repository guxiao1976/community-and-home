// Component test — notice.vue onCommunitySwitch handles the 10015 branch:
// when the backend rejects a community switch with code 10015 it shows a
// specific toast; other errors and success do not toast.
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §2
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

vi.mock('@dcloudio/uni-app', () => ({
  onPullDownRefresh: vi.fn(),
}));

vi.mock('@/api/community', () => ({
  getNoticeList: vi.fn().mockResolvedValue({ notices: [], total: '0' }),
  getContacts: vi.fn().mockResolvedValue({ contacts: [] }),
  getLostFoundList: vi.fn().mockResolvedValue({ items: [], total: '0' }),
  getNoticeRoleName: vi.fn(() => ''),
  getNoticeRoleColor: vi.fn(() => '#000000'),
  getContactCategoryName: vi.fn(() => ''),
  getContactCategoryIcon: vi.fn(() => ''),
  getLostFoundTypeName: vi.fn(() => ''),
}));

vi.mock('@/api/user', () => ({
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  setCurrentCommunity: vi.fn(),
}));

import { setCurrentCommunity } from '@/api/user';
import { useCommunityStore } from '@/stores/community';
import NoticePage from './notice.vue';

let pinia: Pinia;

async function mountPage(): Promise<VueWrapper> {
  const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
  await flushPromises();
  return wrapper;
}

function seedCommunities(store: ReturnType<typeof useCommunityStore>) {
  store.addCommunity({ communityId: 'c1', communityName: 'A 小区' });
  store.addCommunity({ communityId: 'c2', communityName: 'B 小区' });
  store.currentCommunityId = 'c1';
}

describe('notice page — onCommunitySwitch 10015 branch', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  it('shows a specific toast when switch fails with code 10015', async () => {
    (setCurrentCommunity as any).mockRejectedValue(
      Object.assign(new Error('目标小区不在数据范围'), { code: 10015 }),
    );
    const wrapper = await mountPage();
    const store = useCommunityStore();
    seedCommunities(store);

    await (wrapper.vm as any).onCommunitySwitch('c2');

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '目标小区不在你的数据范围' }),
    );
    // 当前小区保持不变（权威校验在后端，前端不吞错）
    expect(store.currentCommunityId).toBe('c1');
  });

  it('does not toast for a non-10015 switch error', async () => {
    (setCurrentCommunity as any).mockRejectedValue(
      Object.assign(new Error('服务不可用'), { code: 500 }),
    );
    const wrapper = await mountPage();
    const store = useCommunityStore();
    seedCommunities(store);

    await (wrapper.vm as any).onCommunitySwitch('c2');

    expect(uni.showToast).not.toHaveBeenCalled();
  });

  it('does not toast and updates current community on successful switch', async () => {
    (setCurrentCommunity as any).mockResolvedValue(undefined);
    const wrapper = await mountPage();
    const store = useCommunityStore();
    seedCommunities(store);

    await (wrapper.vm as any).onCommunitySwitch('c2');

    expect(uni.showToast).not.toHaveBeenCalled();
    expect(store.currentCommunityId).toBe('c2');
  });
});
