// Component test — join-choice 页：选小区后的身份分流页
//  - 页顶显示「已加入 XX，请选择下一步」（加入已由 join-community 立即完成）+ 保留社区名/地址
//  - 【填写房号成为业主】→ navigateTo join-residence（pending 小区随行）
//  - 【其他身份认证】→ 设 communityStore.pendingCommunityId（供我的页申请角色）+ switchTab my
//  - 无 pending-join（深链）→ 空态提示，不导航
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { createPinia, setActivePinia, type Pinia } from 'pinia';

vi.mock('@/utils/pending-join', () => ({
  savePendingJoin: vi.fn(),
  readPendingJoin: vi.fn(() => null),
  clearPendingJoin: vi.fn(),
}));

import { readPendingJoin } from '@/utils/pending-join';
import { useCommunityStore } from '@/stores/community';
import JoinChoicePage from './join-choice.vue';

let pinia: Pinia;

const PENDING = { communityId: 'c1', communityName: '幸福小区', address: '幸福路1号' };

async function mountWithPending(value: typeof PENDING | null) {
  (readPendingJoin as any).mockReturnValue(value);
  const wrapper = mount(JoinChoicePage, { global: { plugins: [pinia] } });
  await flushPromises();
  return wrapper;
}

describe('join-choice page — 身份分流', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  it('页顶显示「已加入 XX，请选择下一步」+ 保留社区名/地址', async () => {
    const wrapper = await mountWithPending(PENDING);
    expect(wrapper.find('.header-title').text()).toBe('已加入 幸福小区');
    expect(wrapper.find('.header-sub').text()).toBe('请选择下一步');
    expect(wrapper.find('.header-addr').text()).toBe('幸福路1号');
  });

  it('点击「填写房号成为业主」→ navigateTo join-residence', async () => {
    const wrapper = await mountWithPending(PENDING);
    await wrapper.findAll('.choice-card')[0].trigger('click');
    expect(uni.navigateTo).toHaveBeenCalledWith({ url: '/pages/join-residence/join-residence' });
  });

  it('点击「其他身份认证」→ 设 pendingCommunityId + switchTab 我的页', async () => {
    const wrapper = await mountWithPending(PENDING);
    await wrapper.findAll('.choice-card')[1].trigger('click');
    const communityStore = useCommunityStore();
    expect(communityStore.pendingCommunityId).toBe('c1');
    expect(uni.switchTab).toHaveBeenCalledWith({ url: '/pages/my/my' });
  });

  it('无 pending-join（深链）→ 空态提示，不渲染身份卡片', async () => {
    const wrapper = await mountWithPending(null);
    expect(wrapper.find('.choice-card').exists()).toBe(false);
    expect(wrapper.find('.empty').exists()).toBe(true);
  });
});
