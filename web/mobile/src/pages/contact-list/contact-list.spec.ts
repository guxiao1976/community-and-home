// Component test — contact-list.vue 联络拨号网格（REQ-CLP-1）。
// 有逻辑函数：fetch（getContacts）+ 失败态分支 + 拨号点击 → TDD RED→GREEN。
// SEE: [[verify-api-before-calling]] — 路由 GET /api/community/contacts 已在 graph-context 确认；
//       加载失败明确提示不静默
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

let onLoadCb: ((query: Record<string, unknown>) => void) | null = null;

vi.mock('@dcloudio/uni-app', () => ({
  onLoad: (cb: (query: Record<string, unknown>) => void) => { onLoadCb = cb; },
}));

vi.mock('@/api/community', () => ({
  getContacts: vi.fn(),
  getContactCategoryName: vi.fn((cat: number) => (cat === 1 ? '供水维修' : '') ),
  getContactCategoryIcon: vi.fn((cat: number) => (cat === 1 ? '💧' : '📞')),
}));

import { getContacts } from '@/api/community';
import { useCommunityStore } from '@/stores/community';
import ContactListPage from './contact-list.vue';

let pinia: Pinia;

async function mountPage(): Promise<VueWrapper> {
  const store = useCommunityStore();
  store.currentCommunityId = 'c1';
  const wrapper = mount(ContactListPage, { global: { plugins: [pinia] } });
  onLoadCb?.({});
  await flushPromises();
  return wrapper;
}

function seedContacts() {
  (getContacts as any).mockResolvedValue({
    contacts: [
      { id: 'ct1', community_id: 'c1', category: 1, name: '供水维修', phone: '12345678901', sort_order: 1 },
      { id: 'ct2', community_id: 'c1', category: 7, name: '小区民警', phone: '11000000000', sort_order: 2 },
    ],
  });
}

describe('contact-list — 联络拨号网格（REQ-CLP-1）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (getContacts as any).mockResolvedValue({ contacts: [] });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('以当前小区 id 调用 getContacts 并渲染拨号网格（类别图标 + 名称 + 电话）', async () => {
    seedContacts();
    const wrapper = await mountPage();

    expect(getContacts).toHaveBeenCalledWith('c1');
    const cells = wrapper.findAll('.contact-cell');
    expect(cells.length).toBe(2);
    expect(wrapper.text()).toContain('供水维修');
    expect(wrapper.text()).toContain('12345678901');
    expect(wrapper.text()).toContain('小区民警');
  });

  it('点击联络卡片 → uni.makePhoneCall 拨号', async () => {
    seedContacts();
    const wrapper = await mountPage();

    await wrapper.findAll('.contact-cell')[0].trigger('click');

    expect(uni.makePhoneCall).toHaveBeenCalledWith(
      expect.objectContaining({ phoneNumber: '12345678901' }),
    );
  });

  it('联络数据为空 → 空态「暂无联络信息」', async () => {
    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('暂无联络信息');
  });

  it('列表加载失败 → 明确加载失败提示 + console.error（不静默）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (getContacts as any).mockRejectedValue(new Error('联络列表请求失败'));

    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('加载失败');
    expect(errSpy).toHaveBeenCalled();
  });
});
