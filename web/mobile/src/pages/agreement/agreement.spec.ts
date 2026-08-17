// Unit tests for the registration agreement page (src/pages/agreement/agreement.vue).
// 协议页：展示《社区家园使用协议》正文 + checkbox + 确认注册。
// - 未勾选点确认 → toast「请先阅读并同意使用协议」，不调 register
// - 确认注册：读共享模块 readRegPending() → register → 成功 clearRegPending() + 走 handleAuthSuccess（注册完成自动登录一次）
// - 注册失败 → toast 错误并保留临时数据可重试（不调 clearRegPending）
// - 无临时注册数据进入 → 明确提示注册信息已失效
// - 断言共享契约模块 @/utils/reg-pending（readRegPending/clearRegPending），不再依赖 uni storage magic string
// - A: 注册成功分支 submitting 在 handleAuthSuccess 跳转完成（onCompleted）前保持 true（防双提交）
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §13/§19/§21；注册超时 10002/timeout 分流补录 §24
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

let mockOnLoadCb: (options: Record<string, string>) => void = () => {};

vi.mock('@dcloudio/uni-app', () => ({
  onLoad: (cb: (options: Record<string, string>) => void) => {
    mockOnLoadCb = cb;
  },
}));

vi.mock('@/api/identity', () => ({
  register: vi.fn(),
}));

vi.mock('@/api/user', () => ({
  getUserMemberships: vi.fn(),
  getResidentialAreasByIds: vi.fn(),
  getAppState: vi.fn(),
  setCurrentCommunity: vi.fn(),
}));

vi.mock('@/utils/device', () => ({
  getDeviceId: vi.fn(() => 'dev-1'),
}));

vi.mock('@/utils/reg-pending', () => ({
  REG_PENDING_KEY: 'reg_pending',
  saveRegPending: vi.fn(),
  readRegPending: vi.fn(),
  clearRegPending: vi.fn(),
}));

vi.mock('@common/utils/auth', () => ({
  getAccessToken: vi.fn(() => null),
  getRefreshToken: vi.fn(() => null),
  setTokens: vi.fn(),
  clearTokens: vi.fn(),
  isAuthenticated: vi.fn(() => false),
}));

vi.mock('@/utils/auth-flow', () => ({
  handleAuthSuccess: vi.fn().mockResolvedValue(undefined),
}));

import { register } from '@/api/identity';
import { readRegPending, clearRegPending } from '@/utils/reg-pending';
import { handleAuthSuccess } from '@/utils/auth-flow';
import AgreementPage from './agreement.vue';

let pinia: Pinia;

function seedPending() {
  (readRegPending as any).mockReturnValue({
    phone: '13800138000',
    smsCode: '123456',
    deviceId: 'dev-1',
    nickname: '用户8000',
  });
}

function mountPage(): VueWrapper {
  const wrapper = mount(AgreementPage, { global: { plugins: [pinia] } });
  return wrapper;
}

describe('agreement page — 协议确认注册', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
    (readRegPending as any).mockReturnValue(null);
    mockOnLoadCb = () => {};
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('未勾选 checkbox 点确认 → toast 请先阅读并同意使用协议，不调 register', async () => {
    seedPending();
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '请先阅读并同意使用协议' }),
    );
    expect(register).not.toHaveBeenCalled();
  });

  it('勾选后确认注册 → readRegPending + register 正确参数 + clearRegPending + 自动登录', async () => {
    seedPending();
    (register as any).mockResolvedValue({
      accessToken: 'at',
      refreshToken: 'rt',
      expiresAt: 123,
      userId: 'u1',
    });
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    (wrapper.vm as any).toggleAgreed();
    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    // 注册数据来自共享模块 readRegPending（不再读 uni storage）
    expect(readRegPending).toHaveBeenCalled();
    expect(register).toHaveBeenCalledWith({
      phone: '13800138000',
      smsCode: '123456',
      nickname: '用户8000',
      deviceId: 'dev-1',
    });
    // 注册成功清除临时数据
    expect(clearRegPending).toHaveBeenCalledTimes(1);
    // 注册完成自动登录一次
    expect(handleAuthSuccess).toHaveBeenCalledTimes(1);
  });

  it('A: 确认注册成功 → submitting 在 handleAuthSuccess 跳转完成前保持 true，onCompleted 回调后才复位', async () => {
    seedPending();
    (register as any).mockResolvedValue({
      accessToken: 'at',
      refreshToken: 'rt',
      expiresAt: 123,
      userId: 'u1',
    });
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    (wrapper.vm as any).toggleAgreed();
    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    expect(handleAuthSuccess).toHaveBeenCalledTimes(1);
    // onCompleted 作为第二参数传入（跳转后回调），此时尚未触发
    expect(handleAuthSuccess).toHaveBeenCalledWith(
      expect.objectContaining({ accessToken: 'at' }),
      expect.objectContaining({ onCompleted: expect.any(Function) }),
    );
    // 防双提交窗口：跳转完成前 submitting 必须保持 true
    expect((wrapper.vm as any).submitting).toBe(true);

    // 模拟跳转完成 → onCompleted 触发 → submitting 复位
    const onCompleted = (handleAuthSuccess as any).mock.calls[0][1]
      .onCompleted as () => void;
    onCompleted();
    await flushPromises();
    expect((wrapper.vm as any).submitting).toBe(false);
  });

  it('注册失败 → toast 错误并保留临时数据可重试（不调 clearRegPending）', async () => {
    const errSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    seedPending();
    (register as any).mockRejectedValue(new Error('手机号已被注册'));
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    (wrapper.vm as any).toggleAgreed();
    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    expect(errSpy).toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '注册失败，请重试' }),
    );
    // 临时数据保留，可重试
    expect(clearRegPending).not.toHaveBeenCalled();
    expect(handleAuthSuccess).not.toHaveBeenCalled();
  });

  it('手机号已注册(10002) → 清临时数据 + 回登录页直接登录', async () => {
    seedPending();
    (register as any).mockRejectedValue(Object.assign(new Error('手机号已注册'), { code: 10002 }));
    (uni as any).navigateBack = vi.fn();
    vi.useFakeTimers();
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    (wrapper.vm as any).toggleAgreed();
    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '该手机号已注册，请直接登录' }),
    );
    expect(clearRegPending).toHaveBeenCalled();
    vi.runAllTimers();
    expect((uni as any).navigateBack).toHaveBeenCalledWith({ delta: 1 });
    expect(handleAuthSuccess).not.toHaveBeenCalled();
    vi.useRealTimers();
  });

  it('注册超时（timeout）→ 提示账号可能已创建、保留数据可重试', async () => {
    seedPending();
    (register as any).mockRejectedValue(new Error('timeout of 30000ms exceeded'));
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    (wrapper.vm as any).toggleAgreed();
    await (wrapper.vm as any).confirmRegister();
    await flushPromises();

    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '注册超时，账号可能已创建；请重试或返回登录' }),
    );
    expect(clearRegPending).not.toHaveBeenCalled();
    expect(handleAuthSuccess).not.toHaveBeenCalled();
  });

  it('无临时注册数据进入 → readRegPending 返回 null 并明确提示注册信息已失效', async () => {
    const wrapper = mountPage();
    mockOnLoadCb({});
    await flushPromises();

    expect(readRegPending).toHaveBeenCalled();
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '注册信息已失效，请重新登录' }),
    );
  });
});
