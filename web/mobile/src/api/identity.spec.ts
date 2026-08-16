// Unit tests for src/api/identity.ts — logout (退出登录).
// logout 为纯接线（POST body 透出 deviceId/kickAllDevices），按分诊「字段映射类」只需测试绿，
// 无需 RED 摘录（行为断言：request.post 收到正确路径与 body）。
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/utils/request', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

import request from '@/utils/request';
import { logout } from '@/api/identity';

describe('identity API — logout', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('POSTs deviceId + kickAllDevices=false to /api/auth/logout', async () => {
    (request.post as any).mockResolvedValue({});
    await logout('dev-1');

    expect(request.post).toHaveBeenCalledWith('/api/auth/logout', {
      deviceId: 'dev-1',
      kickAllDevices: false,
    });
  });

  it('kickAllDevices=true 时 body 携带 kickAllDevices: true', async () => {
    (request.post as any).mockResolvedValue({});
    await logout('dev-1', true);

    expect(request.post).toHaveBeenCalledWith('/api/auth/logout', {
      deviceId: 'dev-1',
      kickAllDevices: true,
    });
  });
});
