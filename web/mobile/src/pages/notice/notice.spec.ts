// Component test — notice.vue onCommunitySwitch handles the 10015 branch:
// when the backend rejects a community switch with code 10015 it shows a
// specific toast; other errors and success do not toast.
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §2
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
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

import { setCurrentCommunity, getUserMemberships } from '@/api/user';
import { getNoticeList, getContacts, getLostFoundList } from '@/api/community';
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

describe('notice page — fetch 静默 catch 消除（REQ-P1-ERR）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getContacts as any).mockResolvedValue({ contacts: [] });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    (getUserMemberships as any).mockResolvedValue([
      { community_id: 'c1', community_name: 'A 小区' },
      { community_id: 'c2', community_name: 'B 小区' },
    ]);
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('getLostFoundList 失败 → 寻失 toast + console.error', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getLostFoundList as any).mockRejectedValue(new Error('寻失请求失败'));

    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '寻失加载失败', icon: 'none' }),
    );
    expect(errSpy).toHaveBeenCalled();
  });

  it('getNoticeList 失败 → 通知 toast + console.error', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('通知请求失败'));

    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '通知加载失败', icon: 'none' }),
    );
    expect(errSpy).toHaveBeenCalled();
  });

  it('getContacts 失败 → 联络 toast + console.error', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getContacts as any).mockRejectedValue(new Error('联络请求失败'));

    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '联络加载失败', icon: 'none' }),
    );
    expect(errSpy).toHaveBeenCalled();
  });

  it('三请求并发全部失败 → toast ≥1 次 + console.error 恰好 3 次 + 页面不崩', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('通知请求失败'));
    (getContacts as any).mockRejectedValue(new Error('联络请求失败'));
    (getLostFoundList as any).mockRejectedValue(new Error('寻失请求失败'));

    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalled();
    expect(errSpy).toHaveBeenCalledTimes(3);
    // 页面不崩：loading 复位 + 内容区（空态）渲染而非骨架屏
    expect((wrapper.vm as any).loading).toBe(false);
    expect(wrapper.text()).toContain('暂无通知公告');
    expect(wrapper.text()).toContain('暂无寻失信息');
  });

  it('局部失败（通知失败 + 其余成功）→ 成功区块数据仍渲染', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('通知请求失败'));
    (getContacts as any).mockResolvedValue({
      contacts: [{ id: 'ct1', community_id: 'c1', category: 1, name: '供水维修', phone: '12345678901', sort_order: 1 }],
    });
    (getLostFoundList as any).mockResolvedValue({
      items: [{ id: 'lf1', community_id: 'c1', type: 1, title: '丢失的钥匙', description: '', image_urls: [], contact_phone: '13800000000', status: 1, publisher_id: 'u1', created_at: 1234567890 }],
      total: '1',
    });

    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(wrapper.text()).toContain('12345678901');
    expect(wrapper.text()).toContain('丢失的钥匙');
  });

  it('snake_case 字段渲染：通知 created_at 回退 + 寻失 image_urls[0]', async () => {
    // SEE: [[snake-camel-field-mismatch]] — mock 镜像后端 types.go JSON tag，验证 camelCase→snake_case 后渲染不坏
    (getNoticeList as any).mockResolvedValue({
      notices: [{
        id: 'n1', community_id: 'c1', title: '停水通知', content: '明日停水',
        role: 1, publisher: '物业', publisher_id: 'p1', is_pinned: false,
        published_at: 0, created_at: 1700000000, updated_at: 1700000000, attachments: [],
      }],
      total: '1',
    });
    (getLostFoundList as any).mockResolvedValue({
      items: [{
        id: 'lf2', community_id: 'c1', type: 1, title: '丢失的手表', description: '',
        image_urls: ['https://example.com/a.jpg'], contact_phone: '13800000000',
        status: 1, publisher_id: 'u1', created_at: 1700000000,
      }],
      total: '1',
    });

    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    // 通知/寻失标题均从 snake_case mock 渲染
    expect(wrapper.text()).toContain('停水通知');
    expect(wrapper.text()).toContain('丢失的手表');
    // published_at=0 时回退 created_at 渲染时间（year=2023）
    expect(wrapper.text()).toMatch(/2023/);
    // 寻失缩略图取 image_urls[0]
    const img = wrapper.find('.lost-found-image');
    expect(img.exists()).toBe(true);
    expect(img.attributes('src')).toBe('https://example.com/a.jpg');
  });

  it('成功场景 → 无错误 toast', async () => {
    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).not.toHaveBeenCalled();
  });
});
