// Unit tests for src/utils/pending-join.ts — 加入小区流程跨页一次性「待加入小区」数据的唯一契约模块
// - save→read 往返（模块级内存态为主载体）
// - H5 镜像到 sessionStorage（{data, expiresAt}），绝不 localStorage
// - TTL 过期 → 返回 null 并清除镜像
// - clear → 内存 + 镜像一并清除
// - 空数据（未保存）→ null
// SEE: [[tdd-red-evidence-requires-fail-excerpt]] — RED 摘录见 _tdd_evidence.md
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  PENDING_JOIN_KEY,
  savePendingJoin,
  readPendingJoin,
  clearPendingJoin,
  type PendingJoin,
} from './pending-join';

const PENDING: PendingJoin = {
  communityId: 'c1',
  communityName: '幸福小区',
  address: '幸福路1号',
};

describe('pending-join — 跨页一次性待加入小区契约', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    clearPendingJoin();
    window.sessionStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('save→read 往返：内存态返回同一数据', () => {
    savePendingJoin(PENDING);
    expect(readPendingJoin()).toEqual(PENDING);
  });

  it('membershipId 可选字段：join-community 立即加入回填后随 save→read 往返', () => {
    // 新模型：加入=立即建 membership，join-community 把 membership.id 回填 pending-join，
    // join-residence 绑定房号（bindResidence）时优先读取。
    savePendingJoin({ ...PENDING, membershipId: 'm1' });
    expect(readPendingJoin()).toEqual({ ...PENDING, membershipId: 'm1' });
  });

  it('未填 membershipId（深链/旧数据）→ read 返回对象不含该字段', () => {
    savePendingJoin(PENDING);
    expect(readPendingJoin()).toEqual(PENDING);
    expect(readPendingJoin()).not.toHaveProperty('membershipId');
  });

  it('H5 镜像写入 sessionStorage 且带 {data, expiresAt}，绝不触碰 localStorage', () => {
    const lsSet = vi.spyOn(window.localStorage, 'setItem');
    const lsGet = vi.spyOn(window.localStorage, 'getItem');
    const lsRemove = vi.spyOn(window.localStorage, 'removeItem');

    savePendingJoin(PENDING);

    const raw = window.sessionStorage.getItem(PENDING_JOIN_KEY);
    expect(raw).toBeTruthy();
    const parsed = JSON.parse(raw as string);
    expect(parsed.data).toEqual(PENDING);
    expect(typeof parsed.expiresAt).toBe('number');
    expect(parsed.expiresAt).toBeGreaterThan(Date.now());
    expect(lsSet).not.toHaveBeenCalled();
    expect(lsGet).not.toHaveBeenCalled();
    expect(lsRemove).not.toHaveBeenCalled();
  });

  it('TTL 30 分钟过期 → 返回 null 并清除镜像', () => {
    vi.useFakeTimers();
    savePendingJoin(PENDING);
    expect(readPendingJoin()).toEqual(PENDING);

    vi.advanceTimersByTime(30 * 60 * 1000 + 1);
    expect(readPendingJoin()).toBeNull();
    expect(window.sessionStorage.getItem(PENDING_JOIN_KEY)).toBeNull();
  });

  it('clear → 内存态与 session 镜像一并清除，之后读返回 null', () => {
    savePendingJoin(PENDING);
    clearPendingJoin();
    expect(readPendingJoin()).toBeNull();
    expect(window.sessionStorage.getItem(PENDING_JOIN_KEY)).toBeNull();
  });

  it('空数据（未保存）→ 返回 null', () => {
    expect(readPendingJoin()).toBeNull();
  });
});
