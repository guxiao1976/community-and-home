// Component test — join-community page 选小区流程改造：
// 选中小区后不再弹「自有/租住 + 楼单元房号」modal，改为把 {id, name, address}
// 存入 pending-join 共享契约模块 → navigateTo /pages/join-choice/join-choice。
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
}));

vi.mock('@/utils/pending-join', () => ({
  savePendingJoin: vi.fn(),
  readPendingJoin: vi.fn(() => null),
  clearPendingJoin: vi.fn(),
}));

import { savePendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';
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

describe('join-community page — 选小区分流到 join-choice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    (savePendingJoin as any).mockClear();
  });

  it('点击未加入小区 → 存 pending-join {id,name,address} + navigateTo join-choice', async () => {
    const wrapper = await mountAtSearchStep();
    await wrapper.find('.list-item').trigger('click');

    expect(savePendingJoin).toHaveBeenCalledWith({
      communityId: 'c1',
      communityName: '幸福小区',
      address: '幸福路1号',
    });
    expect(uni.navigateTo).toHaveBeenCalledWith({ url: '/pages/join-choice/join-choice' });
  });

  it('不再弹出 join form modal（.join-form-mask 不存在）', async () => {
    const wrapper = await mountAtSearchStep();
    await wrapper.find('.list-item').trigger('click');

    expect(wrapper.find('.join-form-mask').exists()).toBe(false);
    expect(wrapper.find('.confirm-join-btn').exists()).toBe(false);
  });

  it('点击已加入小区 → 不存 pending、不导航', async () => {
    const wrapper = await mountAtSearchStep();
    const communityStore = useCommunityStore();
    communityStore.addCommunity({ communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' });
    await (wrapper.vm as any).$nextTick();

    await wrapper.find('.list-item').trigger('click');

    expect(savePendingJoin).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
  });

  it('已达上限（3 个小区）→ 显示上限警告，不存 pending、不导航', async () => {
    const wrapper = await mountAtSearchStep();
    const communityStore = useCommunityStore();
    for (let i = 1; i <= 3; i++) {
      communityStore.addCommunity({ communityId: `c${i}`, communityName: `小区${i}` });
    }
    await (wrapper.vm as any).$nextTick();

    await wrapper.find('.list-item').trigger('click');

    expect(savePendingJoin).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect((wrapper.vm as any).showMaxWarning).toBe(true);
  });
});
