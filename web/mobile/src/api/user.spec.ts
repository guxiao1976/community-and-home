// Unit tests for src/api/user.ts — joinCommunity carries community ownership.
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — 本文件 RED 阶段捕获了真实 FAIL 摘录
// （request.post 仅收到 {community_id}，缺 building/unit/room/ownership），见 _tdd_evidence.md §1。
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/utils/request', () => ({
  default: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

import request from '@/utils/request';
import { joinCommunity } from '@/api/user';

function mockMembership(overrides: Record<string, unknown> = {}) {
  return {
    id: 'm1',
    user_id: 'u1',
    community_id: 'c1',
    bind_status: 1,
    building: 1,
    unit: 2,
    room: 301,
    join_time: 0,
    leave_time: 0,
    created_at: 0,
    updated_at: 0,
    ...overrides,
  };
}

describe('joinCommunity', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('POSTs community_id/building/unit/room/ownership to /api/users/communities/join', async () => {
    (request.post as any).mockResolvedValue(mockMembership());
    const membership = await joinCommunity('c1', 1, 2, 301, 1);

    expect(request.post).toHaveBeenCalledWith('/api/users/communities/join', {
      community_id: 'c1',
      building: 1,
      unit: 2,
      room: 301,
      ownership: 1,
    });
    expect(membership).toMatchObject({ community_id: 'c1', building: 1, unit: 2, room: 301 });
  });

  it('sends ownership=2 (RENTED) and keeps Snowflake community_id as string', async () => {
    const snowflakeId = '1234567890123456789';
    (request.post as any).mockResolvedValue(mockMembership({ community_id: snowflakeId, building: 3, unit: 1, room: 502 }));
    await joinCommunity(snowflakeId, 3, 1, 502, 2);

    expect(request.post).toHaveBeenCalledWith('/api/users/communities/join', {
      community_id: snowflakeId,
      building: 3,
      unit: 1,
      room: 502,
      ownership: 2,
    });
  });
});
