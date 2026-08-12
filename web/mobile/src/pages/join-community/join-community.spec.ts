// Component test — join-community page collects 自有/租住 (ownership) + building/unit/room
// before calling joinCommunity, and validates them.
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { createPinia, setActivePinia, type Pinia } from 'pinia';

vi.mock('@/api/user', () => ({
  getDivisions: vi.fn().mockResolvedValue([]),
  searchResidentialAreas: vi.fn().mockResolvedValue([]),
  joinCommunity: vi.fn(),
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
}));

import { joinCommunity } from '@/api/user';
import JoinCommunityPage from './join-community.vue';

const area = { id: 'c1', name: '幸福小区', address: '幸福路1号' };

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

async function openJoinForm(wrapper: VueWrapper) {
  await wrapper.find('.list-item').trigger('click');
  await (wrapper.vm as any).$nextTick();
}

async function fillValidForm(wrapper: VueWrapper, ownershipIndex = 0) {
  // ownership radio (0=自有, 1=租住)
  await wrapper.findAll('.ownership-option')[ownershipIndex].trigger('click');
  const inputs = wrapper.findAll('.join-form-input');
  await inputs[0].setValue('3'); // building
  await inputs[1].setValue('1'); // unit
  await inputs[2].setValue('502'); // room
}

describe('join-community page — ownership join form', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (joinCommunity as any).mockResolvedValue({
      id: 'm1', user_id: 'u1', community_id: 'c1', bind_status: 1,
      building: 3, unit: 1, room: 502, join_time: 0, leave_time: 0, created_at: 0, updated_at: 0,
    });
  });

  it('opens the join form when clicking a not-joined community', async () => {
    const wrapper = await mountAtSearchStep();
    await openJoinForm(wrapper);
    expect((wrapper.vm as any).showJoinForm).toBe(true);
    expect(joinCommunity).not.toHaveBeenCalled();
  });

  it('blocks confirm when ownership is not selected', async () => {
    const wrapper = await mountAtSearchStep();
    await openJoinForm(wrapper);
    const inputs = wrapper.findAll('.join-form-input');
    await inputs[0].setValue('3');
    await inputs[1].setValue('1');
    await inputs[2].setValue('502');
    await wrapper.find('.confirm-join-btn').trigger('click');

    expect(joinCommunity).not.toHaveBeenCalled();
    expect((wrapper.vm as any).joinFormErrors.ownership).toBeTruthy();
  });

  it('calls joinCommunity(communityId, building, unit, room, OWNED) for 自有 join', async () => {
    const wrapper = await mountAtSearchStep();
    await openJoinForm(wrapper);
    await fillValidForm(wrapper, 0);

    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 1);
  });

  it('calls joinCommunity with ownership=2 (RENTED) for 租住 join', async () => {
    const wrapper = await mountAtSearchStep();
    await openJoinForm(wrapper);
    await fillValidForm(wrapper, 1);

    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect(joinCommunity).toHaveBeenCalledWith('c1', 3, 1, 502, 2);
  });

  it('shows success card and records the community after a valid join', async () => {
    const wrapper = await mountAtSearchStep();
    await openJoinForm(wrapper);
    await fillValidForm(wrapper, 0);
    await wrapper.find('.confirm-join-btn').trigger('click');
    await flushPromises();

    expect((wrapper.vm as any).joinedArea).toEqual(area);
    expect((wrapper.vm as any).showJoinForm).toBe(false);
    expect(wrapper.find('.success-card').exists()).toBe(true);
  });
});
