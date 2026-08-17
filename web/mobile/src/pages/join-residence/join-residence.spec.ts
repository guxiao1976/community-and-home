// Component test — join-residence 页：房号登记独立步骤（新模型）
// 加入已在上一步（join-community 立即建 membership）完成，本页不再调 joinCommunity；
// 确认 → 读 pending.membershipId（无则 getUserMemberships 按 communityId 回退）→
// bindResidence({membership_id, building, unit, room, is_primary:1}) + applyRole(owner/tenant)。
// 复用 join-form（validateJoinForm / joinFormToPayload / OWNERSHIP_OPTIONS）。
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  bindResidence: vi.fn(),
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

import { bindResidence, applyRole, getUserMemberships } from '@/api/user';
import { clearPendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';
import JoinResidencePage from './join-residence.vue';

let pinia: Pinia;

// 形状与 PendingJoin 对齐（membershipId 可选，join-community 立即加入后回填）
type PendingInput = { communityId: string; communityName: string; address: string; membershipId?: string };
const PENDING: PendingInput = { communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' };

async function mountWithPending(pending: PendingInput = PENDING) {
  const { readPendingJoin } = await import('@/utils/pending-join');
  (readPendingJoin as any).mockReturnValue(pending);
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

describe('join-residence page — 房号登记独立步骤（bindResidence + applyRole）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (bindResidence as any).mockResolvedValue({});
    (applyRole as any).mockResolvedValue({});
    uni.setStorageSync('current_community_id', '');
  });

  it('空表单 → 校验错误展示，不调 bindResidence/applyRole', async () => {
    const wrapper = await mountWithPending({ ...PENDING, membershipId: 'm1' });
    await wrapper.find('.confirm-join-btn').trigger('click');

    expect((wrapper.vm as any).joinFormErrors.ownership).toBeTruthy();
    expect((wrapper.vm as any).joinFormErrors.building).toBeTruthy();
    expect(bindResidence).not.toHaveBeenCalled();
    expect(applyRole).not.toHaveBeenCalled();
  });

  it('无 pending-join（深链）→ toast 请先选择小区，不调 API', async () => {
    const { readPendingJoin } = await import('@/utils/pending-join');
    (readPendingJoin as any).mockReturnValue(null);
    const wrapper = mount(JoinResidencePage, { global: { plugins: [pinia] } });
    await flushPromises();

    await wrapper.find('.confirm-join-btn').trigger('click');

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '请先选择小区' }));
    expect(bindResidence).not.toHaveBeenCalled();
    expect(applyRole).not.toHaveBeenCalled();
  });

  it('自有 → 读 pending.membershipId → bindResidence(is_primary:1) + applyRole(owner) + toast 房号登记成功 + 清 pending + switchTab notice', async () => {
    const wrapper = await mountWithPending({ ...PENDING, membershipId: 'm1' });
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(bindResidence).toHaveBeenCalledWith({
      membership_id: 'm1',
      building: '3',
      unit: '1',
      room: '502',
      is_primary: 1,
    });
    expect(applyRole).toHaveBeenCalledWith({ community_id: 'c1', role_code: 'owner' });
    expect(getUserMemberships).not.toHaveBeenCalled();
    expect(clearPendingJoin).toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '房号登记成功' }));
    expect(uni.switchTab).toHaveBeenCalledWith({ url: '/pages/notice/notice' });
  });

  it('租住 → applyRole role_code=tenant', async () => {
    const wrapper = await mountWithPending({ ...PENDING, membershipId: 'm1' });
    await fillValidForm(wrapper, 1);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(applyRole).toHaveBeenCalledWith({ community_id: 'c1', role_code: 'tenant' });
  });

  it('pending 无 membershipId → 回退 getUserMemberships 按 communityId 取 membership.id', async () => {
    (getUserMemberships as any).mockResolvedValue([
      {
        id: 'm99', user_id: 'u1', community_id: 'c1', bind_status: 1,
        building: 0, unit: 0, room: 0, join_time: 0, leave_time: 0, created_at: 0, updated_at: 0,
      },
    ]);
    const wrapper = await mountWithPending(PENDING);
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(getUserMemberships).toHaveBeenCalled();
    expect(bindResidence).toHaveBeenCalledWith(expect.objectContaining({ membership_id: 'm99' }));
    expect(applyRole).toHaveBeenCalledWith({ community_id: 'c1', role_code: 'owner' });
  });

  it('getUserMemberships 找不到匹配成员关系 → toast 未找到小区成员关系，不调 bindResidence/applyRole、不清 pending', async () => {
    (getUserMemberships as any).mockResolvedValue([]);
    const wrapper = await mountWithPending(PENDING);
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '未找到小区成员关系，请重新加入' }));
    expect(bindResidence).not.toHaveBeenCalled();
    expect(applyRole).not.toHaveBeenCalled();
    expect(clearPendingJoin).not.toHaveBeenCalled();
  });

  it('bindResidence 失败 → toast 错误，不调 applyRole、不清 pending、不 switchTab（可重试）', async () => {
    (bindResidence as any).mockRejectedValue(new Error('房号绑定失败'));
    const wrapper = await mountWithPending({ ...PENDING, membershipId: 'm1' });
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '房号绑定失败' }));
    expect(applyRole).not.toHaveBeenCalled();
    expect(clearPendingJoin).not.toHaveBeenCalled();
    expect(uni.switchTab).not.toHaveBeenCalled();
  });
});
