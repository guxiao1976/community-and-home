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
  getLostFoundList: vi.fn().mockResolvedValue({ items: [], total: '0' }),
  getNoticeRoleName: vi.fn(() => ''),
  getNoticeRoleColor: vi.fn(() => '#000000'),
  getLostFoundTypeName: vi.fn(() => ''),
}));

vi.mock('@/api/user', () => ({
  getUserMemberships: vi.fn().mockResolvedValue([]),
  getResidentialAreasByIds: vi.fn().mockResolvedValue([]),
  setCurrentCommunity: vi.fn(),
  getAppState: vi.fn().mockResolvedValue({ current_community_id: '0', updated_at: 0 }),
}));

import { setCurrentCommunity, getUserMemberships, getAppState } from '@/api/user';
import { getNoticeList, getLostFoundList, getNoticeRoleName } from '@/api/community';
import { useCommunityStore } from '@/stores/community';
import NoticePage from './notice.vue';
import dayjs from 'dayjs';

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

// 默认：登录用户已加入 c1/c2 两个小区，首小区 c1 被选中，内容区渲染
function seedMemberships() {
  (getUserMemberships as any).mockResolvedValue([
    { community_id: 'c1', community_name: 'A 小区' },
    { community_id: 'c2', community_name: 'B 小区' },
  ]);
}

// 预置 uni storage 的 current_community_id='c1'：store 初始化即选中 c1，
// loadMemberships 不再变更 → 不触发 watch 的二次 loadAll，getNoticeList 参数与
// console.error 次数均确定可断言（避免共享 storage 跨用例污染）
function seedStoredCommunity() {
  uni.setStorageSync('current_community_id', 'c1');
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

  it('non-10015 switch error → console.error + 通用 toast（不再静默吞错）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (setCurrentCommunity as any).mockRejectedValue(
      Object.assign(new Error('服务不可用'), { code: 500 }),
    );
    const wrapper = await mountPage();
    const store = useCommunityStore();
    seedCommunities(store);

    await (wrapper.vm as any).onCommunitySwitch('c2');

    expect(errSpy).toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '切换小区失败' }),
    );
    // 当前小区保持不变（权威校验在后端，前端不吞错）
    expect(store.currentCommunityId).toBe('c1');
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
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
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

  it('双请求并发全部失败 → toast ≥1 次 + console.error 恰好 2 次 + 页面不崩', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('通知请求失败'));
    (getLostFoundList as any).mockRejectedValue(new Error('寻失请求失败'));

    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalled();
    expect(errSpy).toHaveBeenCalledTimes(2);
    // 页面不崩：loading 复位 + 内容区（标题栏）渲染而非骨架屏
    // 通知为空时不渲染空态块，仅保留标题栏；寻失空态保留
    expect((wrapper.vm as any).loading).toBe(false);
    expect(wrapper.text()).toContain('通知公告');
    expect(wrapper.find('.notice-list').exists()).toBe(false);
    expect(wrapper.text()).toContain('暂无寻失信息');
  });

  it('局部失败（通知失败 + 寻失成功）→ 成功区块数据仍渲染', async () => {
    vi.spyOn(console, 'error').mockImplementation(() => {});
    (getNoticeList as any).mockRejectedValue(new Error('通知请求失败'));
    (getLostFoundList as any).mockResolvedValue({
      items: [{ id: 'lf1', community_id: 'c1', type: 1, title: '丢失的钥匙', description: '', image_urls: [], contact_phone: '13800000000', status: 1, publisher_id: 'u1', created_at: 1234567890 }],
      total: '1',
    });

    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

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
    // published_at=0 时回退 created_at 渲染时间 → YYYY-MM-DD（与组件同一 dayjs 计算，时间区鲁棒）
    const expectedDate = dayjs.unix(1700000000).format('YYYY-MM-DD');
    expect(wrapper.text()).toContain(expectedDate);
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

describe('notice page — 首页通知 30 天窗口参数（REQ-NTW-1/2）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('首页通知以 since_days=30 & page_size=3 调用 getNoticeList', async () => {
    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(getNoticeList).toHaveBeenCalledWith('c1', 1, 3, 30);
  });
});

describe('notice page — 4 功能图标入口（REQ-FE-1/2/3）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('4 个入口按固定顺序渲染：便民联络/物业报修/二手闲置/租房卖房', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const entries = wrapper.findAll('.func-entry');
    expect(entries.length).toBe(4);
    const labels = entries.map(e => e.find('.func-entry-label').text());
    expect(labels).toEqual(['便民联络', '物业报修', '二手闲置', '租房卖房']);
  });

  it('点击便民联络 → uni.navigateTo 到 contact-list（做实跳页）', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const entry = wrapper.findAll('.func-entry')[0];
    await entry.trigger('click');

    expect(uni.navigateTo).toHaveBeenCalledWith(
      expect.objectContaining({ url: '/pages/contact-list/contact-list' }),
    );
    expect(uni.showToast).not.toHaveBeenCalled();
  });

  it.each([
    [1, '物业报修'],
    [2, '二手闲置'],
    [3, '租房卖房'],
  ])('点击 %s 占位入口 → showToast 功能开发中 且不跳转', async (idx, label) => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const entry = wrapper.findAll('.func-entry')[idx as number];
    await entry.trigger('click');

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '功能开发中' }),
    );
    expect(uni.navigateTo).not.toHaveBeenCalled();
  });

  it('未加入小区（hasCommunities=false）时不渲染入口区', async () => {
    (getUserMemberships as any).mockResolvedValue([]);
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(wrapper.findAll('.func-entry').length).toBe(0);
    expect(wrapper.text()).toContain('请先加入小区以查看更多内容');
  });
});

describe('notice page — 通知模块标题栏 + 两行布局（标题全文 + 元信息行）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('跑马灯已移除：页面无 .marquee-bar，替换为标题栏「通知公告」+「更多」', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(wrapper.find('.marquee-bar').exists()).toBe(false);
    expect(wrapper.find('.notice-header').exists()).toBe(true);
    expect(wrapper.text()).toContain('通知公告');
    expect(wrapper.text()).toContain('更多');
  });

  it('通知为空 → 不渲染列表与空态块，仅保留标题栏', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(wrapper.find('.notice-list').exists()).toBe(false);
    expect(wrapper.text()).not.toContain('暂无通知公告');
    // 标题栏仍在
    expect(wrapper.find('.notice-header').exists()).toBe(true);
  });

  it('点击「更多」→ 空态也允许进浏览页（无 notices.length===0 拦截）', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    await (wrapper.vm as any).onMoreNotice();
    expect(uni.navigateTo).toHaveBeenCalledWith(
      expect.objectContaining({ url: '/pages/notice-browse/notice-browse' }),
    );
  });

  it('两行布局：标题全文渲染 + 标题下方发布单位 + 发布日期(YYYY-MM-DD)', async () => {
    const longTitle = '小区停水通知：因管道检修明日上午9点至下午5点停水';
    (getNoticeList as any).mockResolvedValue({
      notices: [{
        id: 'n1', community_id: 'c1', title: longTitle, content: '',
        role: 3, publisher: '物业管理处', publisher_id: 'p1', is_pinned: false,
        published_at: 1700000000, created_at: 1700000000, updated_at: 1700000000, attachments: [],
      }],
      total: '1',
    });
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const card = wrapper.find('.notice-card');
    expect(card.exists()).toBe(true);
    // 标题全文渲染（不再 JS 截断，无省略号）
    const title = card.find('.notice-title');
    expect(title.exists()).toBe(true);
    expect(title.text()).toBe(longTitle);
    expect(title.text()).not.toContain('…');
    // 元信息行：发布单位 + 发布日期（YYYY-MM-DD）
    const meta = card.find('.notice-meta');
    expect(meta.exists()).toBe(true);
    expect(meta.text()).toContain('物业管理处');
    const expectedDate = dayjs.unix(1700000000).format('YYYY-MM-DD');
    expect(meta.text()).toContain(expectedDate);
    // 角色色条保留行首
    expect(card.find('.notice-bar').exists()).toBe(true);
  });

  it('publisher 为空 → 发布单位回退 getNoticeRoleName 角色名', async () => {
    (getNoticeList as any).mockResolvedValue({
      notices: [{
        id: 'n1', community_id: 'c1', title: '电梯检修通知', content: '',
        role: 2, publisher: '', publisher_id: 'p1', is_pinned: false,
        published_at: 1700000000, created_at: 1700000000, updated_at: 1700000000, attachments: [],
      }],
      total: '1',
    });
    (getNoticeRoleName as any).mockReturnValue('业委会');
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const meta = wrapper.find('.notice-meta');
    expect(meta.exists()).toBe(true);
    expect(meta.text()).toContain('业委会');
  });

  it('published_at=0 → 回退 created_at 渲染 YYYY-MM-DD', async () => {
    (getNoticeList as any).mockResolvedValue({
      notices: [{
        id: 'n1', community_id: 'c1', title: '停水通知', content: '',
        role: 1, publisher: '社区', publisher_id: 'p1', is_pinned: false,
        published_at: 0, created_at: 1700000000, updated_at: 1700000000, attachments: [],
      }],
      total: '1',
    });
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    const expectedDate = dayjs.unix(1700000000).format('YYYY-MM-DD');
    expect(wrapper.find('.notice-meta').text()).toContain(expectedDate);
  });

  it('无单行布局残留：.notice-line 与 .notice-role-pill 已移除', async () => {
    (getNoticeList as any).mockResolvedValue({
      notices: [{
        id: 'n1', community_id: 'c1', title: '停水通知', content: '',
        role: 3, publisher: '物业', publisher_id: 'p1', is_pinned: false,
        published_at: 1700000000, created_at: 1700000000, updated_at: 1700000000, attachments: [],
      }],
      total: '1',
    });
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(wrapper.find('.notice-line').exists()).toBe(false);
    expect(wrapper.find('.notice-role-pill').exists()).toBe(false);
    // 两行结构：标题行 + 元信息行均存在
    expect(wrapper.find('.notice-title').exists()).toBe(true);
    expect(wrapper.find('.notice-meta').exists()).toBe(true);
  });
});

describe('notice page — 区块垂直全序（REQ-HL-4）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('固定全序：4功能入口 → 邻里互助占位 → 寻失互助 → 底部广告集中区', async () => {
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    // 邻里互助占位区块存在且为占位文案（不伪造数据）
    const neighbor = wrapper.findAll('.section').find(s => s.text().includes('邻里互助'))!;
    expect(neighbor.text()).toContain('互助功能开发中');

    // 底部广告集中区存在，3 个广告垂直堆叠
    const adSection = wrapper.find('.ad-section');
    expect(adSection.exists()).toBe(true);
    expect(adSection.findAll('.ad-banner').length).toBe(3);

    // 全序断言：func-entries < 邻里互助 < 寻失互助 < 底部广告集中区
    const pos = (a: HTMLElement, b: HTMLElement) =>
      (a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING) !== 0;
    const funcEntries = wrapper.find('.func-entries').element as HTMLElement;
    const neighborEl = neighbor.element as HTMLElement;
    const lostFoundEl = wrapper.findAll('.section').find(s => s.text().includes('寻失互助'))!.element as HTMLElement;
    const adSectionEl = adSection.element as HTMLElement;

    expect(pos(funcEntries, neighborEl)).toBe(true);
    expect(pos(neighborEl, lostFoundEl)).toBe(true);
    expect(pos(lostFoundEl, adSectionEl)).toBe(true);
  });
});

describe('notice page — 首页首载去重（REQ-DBL-1/2/3）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getNoticeList as any).mockResolvedValue({ notices: [], total: '0' });
    (getLostFoundList as any).mockResolvedValue({ items: [], total: '0' });
    seedMemberships();
    seedStoredCommunity();
    // 复位 getAppState：本 describe 的覆写用例会把它改为权威 'c2'，clearAllMocks 不清实现，
    // 若不复位会导致后续用例挂载时 currentCommunityId 也被覆写（跨用例 mock 泄漏）
    (getAppState as any).mockResolvedValue({ current_community_id: '0', updated_at: 0 });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('loadMemberships 覆写 currentCommunityId（C1→C2）→ 通知/寻失接口各仅请求一次（以 C2 为维度，REQ-DBL-1）', async () => {
    // 本地存储陈旧指向 c1，服务端 getAppState 权威返回 c2（在 memberships 内）
    (getAppState as any).mockResolvedValue({ current_community_id: 'c2', updated_at: 0 });

    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    // 覆写触发的那次 watch 被首载守卫忽略 + onMounted 显式单次 loadAll → 各只请求一次
    expect(getNoticeList).toHaveBeenCalledTimes(1);
    expect(getNoticeList).toHaveBeenCalledWith('c2', 1, 3, 30);
    expect(getLostFoundList).toHaveBeenCalledTimes(1);
    expect(getLostFoundList).toHaveBeenCalledWith('c2');
    expect(useCommunityStore().currentCommunityId).toBe('c2');
  });

  it('用户手动切换小区 → watch 正常触发 loadAll（新小区维度单次，REQ-DBL-2）', async () => {
    (setCurrentCommunity as any).mockResolvedValue(undefined);
    const wrapper = mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();
    // 首载已完成（c1 维度一次），清空计数以隔离手动切换触发的加载
    (getNoticeList as any).mockClear();
    (getLostFoundList as any).mockClear();

    await (wrapper.vm as any).onCommunitySwitch('c2');
    await flushPromises();

    expect(getNoticeList).toHaveBeenCalledTimes(1);
    expect(getNoticeList).toHaveBeenCalledWith('c2', 1, 3, 30);
    expect(getLostFoundList).toHaveBeenCalledTimes(1);
    expect(getLostFoundList).toHaveBeenCalledWith('c2');
  });

  it('loadMemberships 整体失败 → 不以陈旧 cid 发请求、无 double-load（REQ-DBL-1 降级）', async () => {
    // 本地 storage 陈旧指向 c1，但 memberships 拉取网络失败 → communities 清空
    (getUserMemberships as any).mockRejectedValue(new Error('memberships 接口不可用'));
    const store = useCommunityStore();

    mount(NoticePage, { global: { plugins: [pinia] } });
    await flushPromises();

    expect(getNoticeList).not.toHaveBeenCalled();
    expect(getLostFoundList).not.toHaveBeenCalled();
    expect(store.hasCommunities).toBe(false);
  });
});
