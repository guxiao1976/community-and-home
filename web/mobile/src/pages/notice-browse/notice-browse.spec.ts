// Component test — notice-browse.vue 从单条翻页改为 30 天卡片列表（REQ-NTW-4/5）。
// 有逻辑函数：fetch（since_days=30）+ 失败态分支 + 卡片点击导航 → TDD RED→GREEN。
// SEE: [[frontend-business-rule-hardcode]] — 30 天窗口后端强制，前端不实现窗口过滤
// SEE: [[verify-api-before-calling]] — 加载失败明确提示不静默
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';
import dayjs from 'dayjs';

vi.mock('@/api/community', () => ({
  getNoticeList: vi.fn().mockResolvedValue({ notices: [], total: '0' }),
  getNoticeRoleName: vi.fn((role: number) => (role === 1 ? '社区' : '')),
  getNoticeRoleColor: vi.fn(() => '#B8956A'),
}));

import { getNoticeList } from '@/api/community';
import { useCommunityStore } from '@/stores/community';
import NoticeBrowsePage from './notice-browse.vue';

let pinia: Pinia;

async function mountPage(): Promise<VueWrapper> {
  const store = useCommunityStore();
  store.currentCommunityId = 'c1';
  const wrapper = mount(NoticeBrowsePage, { global: { plugins: [pinia] } });
  await flushPromises();
  return wrapper;
}

function seedNotices() {
  (getNoticeList as any).mockResolvedValue({
    notices: [
      {
        id: 'n1', community_id: 'c1', title: '停水通知', content: '明日停水',
        role: 3, publisher: '物业', publisher_id: 'p1', is_pinned: true,
        published_at: 1720000000, created_at: 1720000000, updated_at: 1720000000,
        attachments: [],
      },
      {
        id: 'n2', community_id: 'c1', title: '邻里活动', content: '周末聚会',
        role: 1, publisher: '社区', publisher_id: 'p2', is_pinned: false,
        published_at: 1719000000, created_at: 1719000000, updated_at: 1719000000,
        attachments: [],
      },
    ],
    total: '2',
  });
}

describe('notice-browse — 30 天卡片列表（REQ-NTW-4/5）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('以 since_days=30 & page_size=50 单请求调用 getNoticeList', async () => {
    seedNotices();
    await mountPage();

    expect(getNoticeList).toHaveBeenCalledWith('c1', 1, 50, 30);
  });

  it('按首页一致卡片契约渲染：role 色条 + 标题 + 时间', async () => {
    seedNotices();
    const wrapper = await mountPage();

    const cards = wrapper.findAll('.browse-card');
    expect(cards.length).toBe(2);
    expect(cards[0].text()).toContain('停水通知');
    expect(cards[1].text()).toContain('邻里活动');
    expect(wrapper.find('.browse-card-bar').exists()).toBe(true);
    expect(wrapper.findAll('.browse-card-bar').length).toBe(2);
    // formatTime 行为断言（REQ-NTW-5 卡片时间契约）：卡片渲染 dayjs.unix(published_at).format('YYYY-MM-DD HH:mm')
    expect(cards[0].text()).toContain(dayjs.unix(1720000000).format('YYYY-MM-DD HH:mm'));
    expect(cards[1].text()).toContain(dayjs.unix(1719000000).format('YYYY-MM-DD HH:mm'));
  });

  it('点卡片 → uni.navigateTo 到 notice-detail?id=...', async () => {
    seedNotices();
    const wrapper = await mountPage();

    await wrapper.findAll('.browse-card')[1].trigger('click');

    expect(uni.navigateTo).toHaveBeenCalledWith(
      expect.objectContaining({ url: '/pages/notice-detail/notice-detail?id=n2' }),
    );
  });

  it('加载失败 → 明确失败提示 + console.error（不静默）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('列表请求失败'));

    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('加载失败');
    expect(errSpy).toHaveBeenCalled();
  });

  it('30 天窗口内无通知 → 空态「暂无通知公告」', async () => {
    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('暂无通知公告');
  });
});
