// Vitest setup file

import { config } from '@vue/test-utils';
import { vi } from 'vitest';

// Mock Element Plus message
vi.mock('element-plus', async () => {
  const actual = await vi.importActual('element-plus');
  return {
    ...actual,
    ElMessage: {
      success: vi.fn(),
      error: vi.fn(),
      warning: vi.fn(),
      info: vi.fn()
    },
    ElMessageBox: {
      confirm: vi.fn().mockResolvedValue('confirm')
    }
  };
});

// Configure Vue Test Utils
config.global.stubs = {
  teleport: true
};
