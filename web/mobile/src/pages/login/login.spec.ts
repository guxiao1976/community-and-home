// Unit tests for login.vue 登录流程改造：
// - 移除协议勾选区，canSubmit 仅依赖手机号+验证码合法
// - loginWithSms 仅当 50001（未注册）→ saveRegPending 暂存 {phone,smsCode,deviceId,nickname} → navigateTo 协议页
// - 其他错误保持现有处理（拦截器已 toast），不暂存、不跳转
// - 断言共享契约模块 @/utils/reg-pending（saveRegPending），不再依赖 uni storage magic string
// - A: 成功分支 submitting 在 handleAuthSuccess 跳转完成（onCompleted）前保持 true（防双提交）
// - B: 错误分支以 err.code 为主判据，msg 字符串仅作 code 缺失时的旧后端兜底
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §4（本次改 §19）
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { mount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { setActivePinia, createPinia, type Pinia } from 'pinia';

vi.mock('@dcloudio/uni-app', () => ({
  onLoad: vi.fn(),
  onUnmounted: vi.fn(),
  onMounted: vi.fn(),
}));

vi.mock('@/api/identity', () => ({
  loginWithSms: vi.fn(),
  sendSmsCode: vi.fn(),
  ensurePublicKey: vi.fn().mockResolvedValue('pub-key'),
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

import { loginWithSms } from '@/api/identity';
import { saveRegPending } from '@/utils/reg-pending';
import { handleAuthSuccess } from '@/utils/auth-flow';
import LoginPage from './login.vue';

let pinia: Pinia;

function mountLogin(): VueWrapper {
  return mount(LoginPage, { global: { plugins: [pinia] } });
}

async function fillAndSubmit(wrapper: VueWrapper, phone: string, code: string) {
  (wrapper.vm as any).phone = phone;
  (wrapper.vm as any).smsCode = code;
  await (wrapper.vm as any).handleSubmit();
  await flushPromises();
}

describe('login page — 登录流程改造', () => {
  beforeEach(() => {
    pinia = createPinia();
    setActivePinia(pinia);
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('协议勾选区已移除：模板无 agreement 块，canSubmit 仅依赖手机号+验证码', () => {
    const wrapper = mountLogin();
    expect(wrapper.find('.agreement').exists()).toBe(false);
    (wrapper.vm as any).phone = '13800138000';
    (wrapper.vm as any).smsCode = '123456';
    expect((wrapper.vm as any).canSubmit).toBe(true);
  });

  it('loginWithSms 返回 50001（未注册）→ saveRegPending 暂存 + navigateTo 协议页', async () => {
    (loginWithSms as any).mockRejectedValue(
      Object.assign(new Error('用户未注册'), { code: 50001 }),
    );
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(saveRegPending).toHaveBeenCalledTimes(1);
    expect(saveRegPending).toHaveBeenCalledWith({
      phone: '13800138000',
      smsCode: '123456',
      deviceId: 'dev-1',
      nickname: '用户8000',
    });
    expect(uni.navigateTo).toHaveBeenCalledWith({ url: '/pages/agreement/agreement' });
    // 未注册分支不触发登录成功流
    expect(handleAuthSuccess).not.toHaveBeenCalled();
  });

  it('loginWithSms 其他错误（非 50001）→ 不暂存、不跳转（拦截器已 toast）', async () => {
    (loginWithSms as any).mockRejectedValue(
      Object.assign(new Error('验证码错误'), { code: 10040 }),
    );
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(saveRegPending).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect(handleAuthSuccess).not.toHaveBeenCalled();
  });

  it('loginWithSms 成功 → 走 handleAuthSuccess（不再本地重复登录流程）', async () => {
    (loginWithSms as any).mockResolvedValue({
      accessToken: 'at',
      refreshToken: 'rt',
      expiresAt: 123,
      userId: 'u1',
    });
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(handleAuthSuccess).toHaveBeenCalledTimes(1);
    expect(saveRegPending).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
  });

  it('A: 登录成功 → submitting 在 handleAuthSuccess 跳转完成前保持 true，onCompleted 回调后才复位', async () => {
    (loginWithSms as any).mockResolvedValue({
      accessToken: 'at',
      refreshToken: 'rt',
      expiresAt: 123,
      userId: 'u1',
    });
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

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

  it('B: code 存在但非 50001（即使 msg 含"未注册"）→ code 为主判据，不进注册流程', async () => {
    (loginWithSms as any).mockRejectedValue(
      Object.assign(new Error('用户未注册'), { code: 10040 }),
    );
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(saveRegPending).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect(handleAuthSuccess).not.toHaveBeenCalled();
    expect((wrapper.vm as any).submitting).toBe(false);
  });

  it('B: code 缺失时回退 msg 兜底（msg 含"未注册"）→ 进入注册流程', async () => {
    (loginWithSms as any).mockRejectedValue(new Error('用户未注册')); // 无 code 字段
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(saveRegPending).toHaveBeenCalledTimes(1);
    expect(uni.navigateTo).toHaveBeenCalledWith({
      url: '/pages/agreement/agreement',
    });
  });

  it('B: code 缺失且 msg 无未注册特征 → 复位 submitting 返回，不进注册流程', async () => {
    (loginWithSms as any).mockRejectedValue(new Error('验证码错误')); // 无 code 字段
    const wrapper = mountLogin();
    await fillAndSubmit(wrapper, '13800138000', '123456');

    expect(saveRegPending).not.toHaveBeenCalled();
    expect(uni.navigateTo).not.toHaveBeenCalled();
    expect((wrapper.vm as any).submitting).toBe(false);
  });
});
