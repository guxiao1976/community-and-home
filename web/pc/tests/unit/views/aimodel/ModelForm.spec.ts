/**
 * ModelForm.vue — API 密钥持久化行为测试
 *
 * 验证修复：新增模型时 API 密钥自动保存到 am_api_key 表。
 * Bug: 之前 api_key 仅在表单中填写，提交后丢失，导致编辑页密钥为空、健康检查失败。
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { mount, flushPromises } from '@vue/test-utils';
import { nextTick } from 'vue';

// ── Mock vue-router ──
const mockPush = vi.fn();
const mockBack = vi.fn();
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: mockPush, back: mockBack }),
  useRoute: () => ({ params: {}, query: {} }),
}));

// ── Mock API module ──
const mockCreateModelConfig = vi.fn();
const mockCreateApiKey = vi.fn();
const mockUpdateModelConfig = vi.fn();
const mockGetModelConfigById = vi.fn();
const mockGetApiKeys = vi.fn();
const mockDeleteApiKey = vi.fn();
const mockUpdateApiKey = vi.fn();
const mockTestModelConnection = vi.fn();
const mockFetchProviderModels = vi.fn();

vi.mock('@/api/aimodel', () => ({
  createModelConfig: (...args: any[]) => mockCreateModelConfig(...args),
  createApiKey: (...args: any[]) => mockCreateApiKey(...args),
  updateModelConfig: (...args: any[]) => mockUpdateModelConfig(...args),
  getModelConfigById: (...args: any[]) => mockGetModelConfigById(...args),
  getApiKeys: (...args: any[]) => mockGetApiKeys(...args),
  deleteApiKey: (...args: any[]) => mockDeleteApiKey(...args),
  updateApiKey: (...args: any[]) => mockUpdateApiKey(...args),
  testModelConnection: (...args: any[]) => mockTestModelConnection(...args),
  fetchProviderModels: (...args: any[]) => mockFetchProviderModels(...args),
}));

vi.mock('@element-plus/icons-vue', () => ({
  ArrowLeft: { template: '<span/>' },
  Plus: { template: '<span/>' },
  Edit: { template: '<span/>' },
  Delete: { template: '<span/>' },
  View: { template: '<span/>' },
  Hide: { template: '<span/>' },
  CircleCheck: { template: '<span/>' },
}));

import ModelForm from '@/views/aimodel/ModelForm.vue';

// ── Helpers ──

/** 共享的 Element Plus 组件 stubs */
const STUBS = {
  'el-card': { template: '<div><slot name="header"/><slot/></div>' },
  'el-form': { template: '<div><slot/></div>' },
  'el-form-item': { template: '<div><slot name="extra"/><slot/></div>' },
  'el-input': {
    template: '<input :value="modelValue" @input="$emit(\'update:modelValue\', $event.target.value)"/>',
    props: ['modelValue', 'type', 'placeholder', 'showPassword'],
    emits: ['update:modelValue'],
  },
  'el-input-number': { template: '<input/>', props: ['modelValue'] },
  'el-select': { template: '<select><slot/></select>', props: ['modelValue'] },
  'el-option': { template: '<option/>', props: ['label', 'value'] },
  'el-button': {
    template: '<button @click="$emit(\'click\')"><slot/></button>',
    props: ['type', 'loading', 'disabled', 'link', 'size'],
    emits: ['click'],
  },
  'el-checkbox-group': { template: '<div><slot/></div>', props: ['modelValue'] },
  'el-checkbox': { template: '<div><slot/></div>', props: ['value', 'label'] },
  'el-tag': { template: '<span><slot/></span>', props: ['type'] },
  'el-table': { template: '<div/>', props: ['data', 'stripe', 'vLoading'] },
  'el-table-column': { template: '<div/>', props: ['prop', 'label', 'minWidth', 'width', 'fixed'] },
  'el-pagination': { template: '<div/>' },'el-dialog': { template: '<div v-if="modelValue"><slot/><slot name="footer"/></div>', props: ['modelValue'] },
  'el-icon': { template: '<span/>' },
  'el-alert': { template: '<div/>' },
};

/**
 * 挂载 ModelForm 并注入 mock formRef.validate。
 * el-form 被 stub 后不提供 validate 方法，需要手动注入。
 */
async function mountModelForm() {
  const wrapper = mount(ModelForm, { global: { stubs: STUBS } });
  await flushPromises();
  await nextTick();

  // 注入 mock formRef：validate 直接回调 true（模拟表单验证通过）
  const vm = wrapper.vm as any;
  vm.formRef = {
    validate: (cb: (valid: boolean) => void) => cb(true),
  };

  return { wrapper, vm };
}

/** 快速填充表单字段 */
function fillForm(vm: any, overrides: Record<string, any> = {}) {
  Object.assign(vm.formData, {
    name: 'test-model',
    display_name: 'Test Model',
    provider: 'openai',
    endpoint: 'https://api.example.com',
    model_type: 'cloud',
    api_key: '',
    ...overrides,
  });
}

describe('ModelForm — API Key Persistence', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockCreateModelConfig.mockResolvedValue({ id: '1234567890123456789' });
    mockCreateApiKey.mockResolvedValue({ id: '9876543210987654321' });
    mockGetApiKeys.mockResolvedValue({ keys: [], total: 0 });
  });

  // ── 核心场景 ──

  it('创建云端模型时，createApiKey 被调用且参数正确', async () => {
    mockCreateModelConfig.mockResolvedValue({ id: '111' });
    mockCreateApiKey.mockResolvedValue({ id: '222' });

    const { vm } = await mountModelForm();
    fillForm(vm, {
      name: 'deepseek-chat',
      provider: 'openai',
      api_key: 'sk-test-key-12345',
    });

    await vm.handleSubmit();
    await flushPromises();

    // 模型创建被调用
    expect(mockCreateModelConfig).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'deepseek-chat', provider: 'openai' })
    );

    // 🔑 核心断言：API 密钥被持久化
    expect(mockCreateApiKey).toHaveBeenCalledWith({
      model_id: '111',
      key_name: 'openai-deepseek-chat-默认',
      api_key: 'sk-test-key-12345',
      description: '创建模型 deepseek-chat 时自动添加',
    });

    expect(mockPush).toHaveBeenCalledWith('/aimodel/models');
  });

  // ── 边界场景 ──

  it('本地模型（无 API 密钥）不调用 createApiKey', async () => {
    mockCreateModelConfig.mockResolvedValue({ id: '333' });
    const { vm } = await mountModelForm();
    fillForm(vm, { name: 'llama3', provider: 'ollama', model_type: 'local', api_key: '' });

    await vm.handleSubmit();
    await flushPromises();

    expect(mockCreateModelConfig).toHaveBeenCalledTimes(1);
    expect(mockCreateApiKey).not.toHaveBeenCalled();
  });

  it('API 密钥创建失败不阻断模型创建，且记录错误', async () => {
    mockCreateModelConfig.mockResolvedValue({ id: '444' });
    mockCreateApiKey.mockRejectedValue(new Error('网络错误'));
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

    const { vm } = await mountModelForm();
    fillForm(vm, { name: 'gpt-4o', provider: 'openai', api_key: 'sk-fake' });

    await vm.handleSubmit();
    await flushPromises();

    // 模型创建成功
    expect(mockCreateModelConfig).toHaveBeenCalledTimes(1);
    // Key 创建被调用
    expect(mockCreateApiKey).toHaveBeenCalledTimes(1);
    // 错误被记录
    expect(consoleSpy).toHaveBeenCalledWith(
      expect.stringContaining('为模型 gpt-4o 创建 API 密钥失败:'),
      expect.any(Error)
    );
    // 仍跳转列表
    expect(mockPush).toHaveBeenCalledWith('/aimodel/models');

    consoleSpy.mockRestore();
  });

  it('空白 API 密钥（含纯空格）不触发创建', async () => {
    mockCreateModelConfig.mockResolvedValue({ id: '555' });
    const { vm } = await mountModelForm();
    fillForm(vm, { name: 'test', provider: 'openai', api_key: '   ' });

    await vm.handleSubmit();
    await flushPromises();

    expect(mockCreateModelConfig).toHaveBeenCalledTimes(1);
    // .trim() 后为空 → 不创建 Key
    expect(mockCreateApiKey).not.toHaveBeenCalled();
  });

  it('批量创建时，每个模型独立创建一条 API Key', async () => {
    mockCreateModelConfig
      .mockResolvedValueOnce({ id: '111' })
      .mockResolvedValueOnce({ id: '222' });

    const { vm } = await mountModelForm();
    fillForm(vm, { name: 'group-model', provider: 'openai', api_key: 'sk-batch' });
    // 模拟勾选了 2 个模型
    vm.selectedModels = ['model-a', 'model-b'];

    await vm.handleSubmit();
    await flushPromises();

    // 2 个模型创建
    expect(mockCreateModelConfig).toHaveBeenCalledTimes(2);
    // 2 条 Key 记录
    expect(mockCreateApiKey).toHaveBeenCalledTimes(2);
    expect(mockCreateApiKey).toHaveBeenNthCalledWith(1, expect.objectContaining({ model_id: '111' }));
    expect(mockCreateApiKey).toHaveBeenNthCalledWith(2, expect.objectContaining({ model_id: '222' }));
  });

  it('表单验证失败不提交', async () => {
    mockCreateModelConfig.mockResolvedValue({ id: '111' });

    const { vm } = await mountModelForm();
    // 注入验证失败的 formRef
    vm.formRef = {
      validate: (cb: (valid: boolean) => void) => cb(false),
    };
    fillForm(vm, { name: 'test', provider: 'openai', api_key: 'sk-test' });

    await vm.handleSubmit();
    await flushPromises();

    // 验证失败 → 不创建模型也不创建密钥
    expect(mockCreateModelConfig).not.toHaveBeenCalled();
    expect(mockCreateApiKey).not.toHaveBeenCalled();
    expect(mockPush).not.toHaveBeenCalled();
  });

  it('createModelConfig 返回空响应时不创建密钥也不崩溃', async () => {
    // 模拟后端返回空 body（极少数异常场景）
    mockCreateModelConfig.mockResolvedValue(null);

    const { vm } = await mountModelForm();
    fillForm(vm, { name: 'test', provider: 'openai', api_key: 'sk-test' });

    await vm.handleSubmit();
    await flushPromises();

    // 模型创建被调用
    expect(mockCreateModelConfig).toHaveBeenCalledTimes(1);
    // res 为 null → 跳过密钥创建（不报错）
    expect(mockCreateApiKey).not.toHaveBeenCalled();
    // 仍然跳转
    expect(mockPush).toHaveBeenCalledWith('/aimodel/models');
  });

  it('编辑模式 API 密钥留空不触发创建也不触发更新', async () => {
    const { vm } = await mountModelForm();
    vm.isEdit = true;
    vm.modelId = '999';
    // 留空 API 密钥
    fillForm(vm, { name: 'existing', provider: 'openai', api_key: '' });

    mockUpdateModelConfig.mockResolvedValue(undefined);

    await vm.handleSubmit();
    await flushPromises();

    // 走 update 分支
    expect(mockUpdateModelConfig).toHaveBeenCalledTimes(1);
    expect(mockCreateModelConfig).not.toHaveBeenCalled();
    // 密钥为空 → 不调用 updateApiKey / createApiKey
    expect(mockCreateApiKey).not.toHaveBeenCalled();
    expect(mockUpdateApiKey).not.toHaveBeenCalled();
  });

  it('编辑模式填写 API 密钥时更新已有密钥', async () => {
    mockUpdateModelConfig.mockResolvedValue(undefined);
    mockUpdateApiKey.mockResolvedValue({});

    const { vm } = await mountModelForm();
    vm.isEdit = true;
    vm.modelId = '999';
    vm.hasExistingKey = true;
    vm.existingKeyId = 'key-001';
    fillForm(vm, { name: 'existing', provider: 'openai', api_key: 'sk-new-key' });

    await vm.handleSubmit();
    await flushPromises();

    // 更新模型配置
    expect(mockUpdateModelConfig).toHaveBeenCalledTimes(1);
    // 更新已有密钥
    expect(mockUpdateApiKey).toHaveBeenCalledWith({ id: 'key-001', api_key: 'sk-new-key' });
    expect(mockCreateApiKey).not.toHaveBeenCalled();
  });

  // ── 编辑模式 canTest ──

  it('编辑模式下 canTest 为 true 当已有密钥', async () => {
    const { vm } = await mountModelForm();
    vm.isEdit = true;
    vm.modelId = '999';
    fillForm(vm, { name: 'existing', provider: 'openai', api_key: '' });
    vm.hasExistingKey = true;

    await nextTick();

    expect(vm.canTest).toBe(true);
  });

  it('编辑模式下 canTest 为 false 当无密钥且未填写', async () => {
    const { vm } = await mountModelForm();
    vm.isEdit = true;
    vm.modelId = '999';
    fillForm(vm, { name: 'existing', provider: 'openai', api_key: '' });
    vm.hasExistingKey = false;

    await nextTick();

    expect(vm.canTest).toBe(false);
  });
});
