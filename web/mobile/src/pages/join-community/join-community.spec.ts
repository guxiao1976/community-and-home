// Component test — join-community page 加入流程重构（新模型）：
// 点「加入小区」→ 立即 joinCommunity(communityId)（无房号）→ addCommunity →
// 把 membership.id 回填 pending-join（新增 membershipId）→ navigateTo join-choice。
// 已加入 → toast 该小区已加入；maxReached → 上限警告；joinCommunity 失败 → toast 错误。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { createPinia, setActivePinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  getDivisions: vi.fn().mockResolvedValue([]),
  searchResidentialAreas: vi.fn().mockResolvedValue([]),
  joinCommunity: vi.fn(),
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  getAppState: vi.fn().mockResolvedValue({ current_community_id: '0', updated_at: 0 }),
}));

vi.mock('@/utils/pending-join', () => ({
  savePendingJoin: vi.fn(),
  readPendingJoin: vi.fn(() => null),
  clearPendingJoin: vi.fn(),
}));

import { joinCommunity } from '@/api/user';
import { savePendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';
import JoinCommunityPage from './join-community.vue';

const area = { id: 'c1', name: '幸福小区', address: '幸福路1号' };

function mockMembership() {
  return {
    id: 'm1', user_id: 'u1', community_id: 'c1', bind_status: 1,
    building: 0, unit: 0, room: 0, join_time: 0, leave_time: 0, created_at: 0, updated_at: 0,
  };
}

async function mountAtSearchStep(): Promise<VueWrapper> {
  const pinia = createPinia();
  setActivePinia(pinia);
  const wrapper = mount(JoinCommunityPage, { global: { plugins: [pinia] } });
  await flushPromises();
  (wrapper.vm as any).step = 4;
  (wrapper.vm as any).areas = [area];
  await (wrapper.vm as any).$nextTick();
  return wrapper;
}

describe('join-community page — 点加入立即建 membership，分流到 join-choice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (savePendingJoin as any).mockClear();
    (joinCommunity as any).mockResolvedValue(mockMembership());
  });

  it('点击未加入小区 → joinCommunity(communityId 无房号) + addCommunity + 存 pending(带 membershipId) + navigateTo join-choice', async () => {
    const wrapper = await mountAtSearchStep();
    await wrapper.find('.list-item').trigger('click');
    await flushPromises();

    expect(joinCommunity).toHaveBeenCalledWith('c1');
    expect(savePendingJoin).toHaveBeenCalledWith({
      communityId: 'c1',
      communityName: '幸福小区',
      address: '幸福路1号',
      membershipId: 'm1',
    });
    const communityStore = useCommunityStore();
    expect(communityStore.communities).toEqual([
      expect.objectContaining({ communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' }),
    ]);
    expect(uni.navigateTo).toHaveBeenCalledWith({ url: '/pages/join-choice/join-choice' });
  });

  it('不再弹出 join form modal（.join-form-mask 不存在）', async () => {
    const wrapper = await mountAtSearchStep();
    await wrapper.find('.list-item').trigger('click');
    await flushPromises();

    expect(wrapper.find('.join-form-mask').exists()).toBe(false);
    expect(wrapper.find('.confirm-join-btn').exists()).toBe(false);
  });

  it('点击已加入小区 → toast 该小区已加入，不调 join、不存 pending、不导航', async () => {
    const wrapper = await mountAtSearchStep();
    const communityStore = useCommunityStore();
    communityStore.addCommunity({ communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' });
    await (wrapper.vm as any).$nextTick();

    await wrapper.find('.list-item').trigger('click');

    expect(joinCommunity).not.toHaveBeenCalled();
    expect(savePendingJoin).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '该小区已加入' }));
  });

  it('已达上限（3 个小区）→ 显示上限警告，不调 join、不存 pending、不导航', async () => {
    const wrapper = await mountAtSearchStep();
    const communityStore = useCommunityStore();
    for (let i = 1; i <= 3; i++) {
      communityStore.addCommunity({ communityId: `c${i}`, communityName: `小区${i}` });
    }
    await (wrapper.vm as any).$nextTick();

    await wrapper.find('.list-item').trigger('click');

    expect(joinCommunity).not.toHaveBeenCalled();
    expect(savePendingJoin).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect((wrapper.vm as any).showMaxWarning).toBe(true);
  });

  it('joinCommunity 失败（如每年最多加入 3 个新小区）→ toast 错误，不存 pending、不导航、不 addCommunity', async () => {
    (joinCommunity as any).mockRejectedValue(new Error('每年最多加入 3 个新小区'));
    const wrapper = await mountAtSearchStep();
    await wrapper.find('.list-item').trigger('click');
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '每年最多加入 3 个新小区' }));
    expect(savePendingJoin).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    const communityStore = useCommunityStore();
    expect(communityStore.communities.length).toBe(0);
  });
});
