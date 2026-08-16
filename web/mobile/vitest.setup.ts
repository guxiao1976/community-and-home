// Vitest global setup — provides a minimal `uni` global so page/store code
// that calls uni.* (getStorageSync, showLoading, showToast, etc.) runs in tests.
import { vi } from 'vitest';

const storage = new Map<string, string>();

const uniStub = {
  getStorageSync: (key: string) => storage.get(key) ?? '',
  setStorageSync: (key: string, value: unknown) => {
    storage.set(key, String(value));
  },
  removeStorageSync: (key: string) => {
    storage.delete(key);
  },
  showLoading: vi.fn(),
  hideLoading: vi.fn(),
  showToast: vi.fn(),
  switchTab: vi.fn(),
  reLaunch: vi.fn(),
  navigateTo: vi.fn(),
  redirectTo: vi.fn(),
  getSystemInfoSync: () => ({ platform: 'h5' }),
  // 附件预览 / 联络拨号（notice-detail / contact-list 测试）
  previewImage: vi.fn(),
  downloadFile: vi.fn(),
  openDocument: vi.fn(),
  makePhoneCall: vi.fn(),
};

// @ts-expect-error — uni is provided by the uni-app runtime in real app; stubbed here for tests
globalThis.uni = uniStub;
