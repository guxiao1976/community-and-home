// Unit tests for the axios response interceptor in src/utils/request.ts.
// The interceptor is a real logic function: it unwraps the "data" field on
// code===0, passes through raw data when there is no code field, and for
// business errors it shows a toast and rejects with an Error that carries the
// business code so callers (e.g. onCommunitySwitch) can branch on it.
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md §3
import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { AxiosResponse } from 'axios';
import request from '@/utils/request';

function okResponse(data: unknown): AxiosResponse {
  return {
    data,
    status: 200,
    statusText: 'OK',
    headers: {},
    config: {} as never,
  };
}

const mockAdapter = vi.fn();

describe('request response interceptor', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    request.defaults.adapter = mockAdapter as never;
  });

  it('unwraps the data field when code === 0', async () => {
    mockAdapter.mockResolvedValue(okResponse({ code: 0, data: { notices: [] } }));

    const res = await request.get('/api/community/notices');

    expect(res).toEqual({ notices: [] });
    expect(uni.showToast).not.toHaveBeenCalled();
  });

  it('returns whole response when code === 0 and no data field present', async () => {
    mockAdapter.mockResolvedValue(okResponse({ code: 0 }));

    const res = await request.get('/api/x');

    expect(res).toEqual({ code: 0 });
  });

  it('returns raw data when response has no code field', async () => {
    mockAdapter.mockResolvedValue(okResponse({ list: [{ id: '1' }] }));

    const res = await request.get('/api/masterdata/divisions');

    expect(res).toEqual({ list: [{ id: '1' }] });
    expect(uni.showToast).not.toHaveBeenCalled();
  });

  it('rejects with an Error carrying the business code for business errors', async () => {
    mockAdapter.mockResolvedValue(okResponse({ code: 10015, msg: '目标小区不在数据范围' }));

    const err = await request.put('/api/users/me/current-community').catch((e) => e);

    expect(err).toBeInstanceOf(Error);
    expect(err.code).toBe(10015);
    expect(err.message).toBe('目标小区不在数据范围');
    expect(uni.showToast).toHaveBeenCalledWith(
      expect.objectContaining({ title: '目标小区不在数据范围' }),
    );
  });

  it('leaves err.code undefined when code is not a number', async () => {
    mockAdapter.mockResolvedValue(okResponse({ code: '50007', msg: '端限制' }));

    const err = await request.get('/api/x').catch((e) => e);

    expect(err.code).toBeUndefined();
    expect(uni.showToast).toHaveBeenCalled();
  });
});
