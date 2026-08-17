// Component test — join-residence 页：业主路径填房号加入
// 复用 join-form（validateJoinForm / joinFormToPayload / OWNERSHIP_OPTIONS）。
// 确认加入 → joinCommunity(communityId, building, unit, room, ownership)
//   → applyRole({community_id, role_code:'owner'}) → addCommunity → 清 pending-join → 提示加入成功。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  joinCommunity: vi.fn(),
  applyRole: vi.fn(),
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  setCurrentCommunity: vi.fn(),
  getAppState: vi.fn().mockResolvedValue({ current_community_id: '0', updated_at: 0 }),
}));

vi.mock('@/utils/pending-join', () => ({
  savePendingJoin: vi.fn(),
  readPendingJoin: vi.fn(() => null),
  clearPendingJoin: vi.fn(),
}));

import { joinCommunity, applyRole } from '@/api/user';
import { clearPendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';
import JoinResidencePage from './join-residence.vue';

let pinia: Pinia;

const PENDING = { communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' };

async function mountWithPending() {
  const { readPendingJoin } = await import('@/utils/pending-join');
  (readPendingJoin as any).mockReturnValue(PENDING);
  const wrapper = mount(JoinResidencePage, { global: { plugins: [pinia] } });
  await flushPromises();
  return wrapper;
}

async function fillValidForm(wrapper: any, ownershipIndex = 0) {
  await wrapper.findAll('.ownership-option')[ownershipIndex].trigger('click');
  const inputs = wrapper.findAll('.join-form-input');
  await inputs[0].setValue('3');
  await inputs[1].setValue('1');
  await inputs[2].setValue('502');
}

describe('join-residence page — 业主路径加入', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (joinCommunity as any).mockResolvedValue({
      id: 'm1', user_id: 'u1', community_id: 'c1', bind_status: 1,
      building: 3, unit: 1, room: 502, join_time: 0, leave_time: 0, created_at: 0, updated_at: 0,
    });
    (applyRole as any).mockResolvedValue({});
    uni.setStorageSync('current_community_id', '');
  });

  it('空表单 → 校验错误展示，不调 API', async () => {
    const wrapper = await mountWithPending();
    await wrapper.find('.confirm-join-btn').trigger('click');

    expect((wrapper.vm as any).joinFormErrors.ownership).toBeTruthy();
    expect((wrapper.vm as any).joinFormErrors.building).toBeTruthy();
    expect(joinCommunity).not.toHaveBeenCalled();
    expect(applyRole).not.toHaveBeenCalled();
  });

  it('无 pending-join（深链）→ toast 请先选择小区，不调 API', async () => {
    const { readPendingJoin } = await import('@/utils/pending-join');
    (readPendingJoin as any).mockReturnValue(null);
    const wrapper = mount(JoinResidencePage, { global: { plugins: [pinia] } });
    await flushPromises();

    await wrapper.find('.confirm-join-btn').trigger('click');

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '请先选择小区' }));
    expect(joinCommunity).not.toHaveBeenCalled();
  });

  it('自有（OWNED=1）→ joinCommunity(ownership=1) + addCommunity + 清 pending + 提示成功，不重复申请角色', async () => {
    // 后端 JoinCommunity 已按权属自动授权 owner/tenant，前端不再 applyRole('owner')（避免租住被误授 owner）。
    const wrapper = await mountWithPending();
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 1);
    expect(applyRole).not.toHaveBeenCalled();
    const communityStore = useCommunityStore();
    expect(communityStore.communities).toEqual([
      expect.objectContaining({ communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' }),
    ]);
    expect(clearPendingJoin).toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '加入成功' }));
    expect(uni.switchTab).toHaveBeenCalledWith({ url: '/pages/notice/notice' });
  });

  it('租住（RENTED=2）→ joinCommunity 传 ownership=2', async () => {
    const wrapper = await mountWithPending();
    await fillValidForm(wrapper, 1);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 2);
  });

  it('joinCommunity 失败 → toast 错误，不 addCommunity、不调 applyRole', async () => {
    (joinCommunity as any).mockRejectedValue(new Error('加入失败，请稍后重试'));
    const wrapper = await mountWithPending();
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '加入失败，请稍后重试' }));
    expect(applyRole).not.toHaveBeenCalled();
    const communityStore = useCommunityStore();
    expect(communityStore.communities.length).toBe(0);
  });
});
