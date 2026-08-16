// Component test — notice-detail.vue 附件预览 file_type 白名单分发（REQ-NDP-2/3/4）。
// 有逻辑函数：附件点击分发分支（图片/文档/空 file_url）+ 详情加载失败态 → TDD RED→GREEN。
// mock 策略：mock @/utils/request 的 get（网络层），保留真实 @/api/community
// （含 isImageAttachment 白名单谓词），避免 importActual 展开与白名单复制。
// SEE: [[frontend-business-rule-hardcode]] — file_type 白名单与 file-service guard/magic.go 对齐
// SEE: [[verify-api-before-calling]] — 加载/预览失败明确提示不静默
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

let onLoadCb: ((query: Record<string, unknown>) => void) | null = null;

vi.mock('@dcloudio/uni-app', () => ({
  onLoad: (cb: (query: Record<string, unknown>) => void) => { onLoadCb = cb; },
}));

vi.mock('@/utils/request', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}));

import request from '@/utils/request';
import NoticeDetailPage from './notice-detail.vue';

let pinia: Pinia;

async function mountPage(): Promise<VueWrapper> {
  const wrapper = mount(NoticeDetailPage, { global: { plugins: [pinia] } });
  // 触发页面 onLoad 读取 query.id 拉取详情
  onLoadCb?.({ id: 'n1' });
  await flushPromises();
  return wrapper;
}

function makeNotice(attachments: unknown[]) {
  (request.get as any).mockResolvedValue({
    notice: {
      id: 'n1', community_id: 'c1', title: '停水通知', content: '明日停水',
      role: 3, publisher: '物业', publisher_id: 'p1', is_pinned: true,
      published_at: 1720000000, created_at: 1720000000, updated_at: 1720000000,
      attachments,
    },
  });
}

describe('notice-detail — 附件预览 file_type 白名单分发（REQ-NDP-2/3/4）', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('图片附件（file_type=png）点击 → uni.previewImage 全屏预览', async () => {
    makeNotice([{ id: 'att1', file_id: 'f1', file_type: 'png', file_name: 'a.png', file_url: 'https://cdn/x.png', file_size: 1024 }]);
    const wrapper = await mountPage();

    await wrapper.find('.attachment-item').trigger('click');

    expect(uni.previewImage).toHaveBeenCalledWith(
      expect.objectContaining({ urls: ['https://cdn/x.png'] }),
    );
    expect(uni.downloadFile).not.toHaveBeenCalled();
  });

  it('文档附件（file_type=pdf）点击 → uni.downloadFile + 成功 openDocument', async () => {
    makeNotice([{ id: 'att2', file_id: 'f2', file_type: 'pdf', file_name: 'b.pdf', file_url: 'https://cdn/b.pdf', file_size: 2048 }]);
    const wrapper = await mountPage();

    await wrapper.find('.attachment-item').trigger('click');

    expect(uni.downloadFile).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://cdn/b.pdf' }),
    );
    expect(uni.previewImage).not.toHaveBeenCalled();
    // 触发 downloadFile 成功回调 → openDocument
    const successCb = (uni.downloadFile as any).mock.calls[0][0].success;
    successCb({ statusCode: 200, tempFilePath: '/tmp/b.pdf' });
    expect(uni.openDocument).toHaveBeenCalledWith(
      expect.objectContaining({ filePath: '/tmp/b.pdf' }),
    );
  });

  it('图片附件 file_url 为空 → toast 预览失败，不降级文档打开器', async () => {
    makeNotice([{ id: 'att3', file_id: '', file_type: 'png', file_name: 'c.png', file_url: '', file_size: 0 }]);
    const wrapper = await mountPage();

    await wrapper.find('.attachment-item').trigger('click');

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '预览失败' }),
    );
    expect(uni.previewImage).not.toHaveBeenCalled();
    expect(uni.downloadFile).not.toHaveBeenCalled();
  });

  it('文档附件 file_url 为空或下载失败 → toast 附件打开失败', async () => {
    makeNotice([{ id: 'att4', file_id: '', file_type: 'pdf', file_name: 'd.pdf', file_url: '', file_size: 0 }]);
    const wrapper = await mountPage();

    await wrapper.find('.attachment-item').trigger('click');

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '附件打开失败' }),
    );
    expect(uni.downloadFile).not.toHaveBeenCalled();
  });

  it('无法识别 file_type → 按文档处理（不崩溃，不直接页内跳转）', async () => {
    makeNotice([{ id: 'att5', file_id: 'f5', file_type: 'application/octet-stream', file_name: 'e.bin', file_url: 'https://cdn/e.bin', file_size: 512 }]);
    const wrapper = await mountPage();

    await wrapper.find('.attachment-item').trigger('click');

    expect(uni.downloadFile).toHaveBeenCalledWith(
      expect.objectContaining({ url: 'https://cdn/e.bin' }),
    );
    expect(uni.previewImage).not.toHaveBeenCalled();
  });

  it('无附件（空数组）→ 不渲染附件区', async () => {
    makeNotice([]);
    const wrapper = await mountPage();

    expect(wrapper.find('.detail-attachments').exists()).toBe(false);
  });

  it('详情加载失败 → 明确失败态「加载失败」+ console.error（不静默）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    (request.get as any).mockRejectedValue(new Error('详情请求失败'));

    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('加载失败');
    expect(errSpy).toHaveBeenCalled();
  });

  it('详情不存在（notice 为 null）→ 通知不存在', async () => {
    (request.get as any).mockResolvedValue({ notice: null });

    const wrapper = await mountPage();

    expect(wrapper.text()).toContain('通知不存在');
  });
});
