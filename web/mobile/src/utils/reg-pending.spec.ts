// Unit tests for src/utils/reg-pending.ts — 登录协议流程跨页一次性注册数据的唯一契约模块
// - save→read 往返（模块级内存态为主载体）
// - H5 镜像到 sessionStorage（{data, expiresAt}，TTL 5 分钟），绝不 localStorage
// - TTL 过期 → 返回 null 并清除镜像
// - clear → 内存 + 镜像一并清除
// - 空数据（未保存）→ null
// - localStorage 全程零调用（消灭持久化残留的根因）
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  REG_PENDING_KEY,
  saveRegPending,
  readRegPending,
  clearRegPending,
  type RegPending,
} from './reg-pending';

const PENDING: RegPending = {
  phone: '13800138000',
  smsCode: '123456',
  deviceId: 'dev-1',
  nickname: '用户8000',
};

describe('reg-pending — 跨页一次性注册数据契约', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    clearRegPending();
    window.sessionStorage.clear();
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.restoreAllMocks();
  });

  it('save→read 往返：内存态返回同一数据', () => {
    saveRegPending(PENDING);
    expect(readRegPending()).toEqual(PENDING);
  });

  it('H5 镜像写入 sessionStorage 且带 {data, expiresAt}，绝不触碰 localStorage', () => {
    const lsSet = vi.spyOn(window.localStorage, 'setItem');
    const lsGet = vi.spyOn(window.localStorage, 'getItem');
    const lsRemove = vi.spyOn(window.localStorage, 'removeItem');

    saveRegPending(PENDING);

    const raw = window.sessionStorage.getItem(REG_PENDING_KEY);
    expect(raw).toBeTruthy();
    const parsed = JSON.parse(raw as string);
    expect(parsed.data).toEqual(PENDING);
    expect(typeof parsed.expiresAt).toBe('number');
    expect(parsed.expiresAt).toBeGreaterThan(Date.now());
    // 一次性验证码等敏感数据绝不落 localStorage
    expect(lsSet).not.toHaveBeenCalled();
    expect(lsGet).not.toHaveBeenCalled();
    expect(lsRemove).not.toHaveBeenCalled();
  });

  it('TTL 5 分钟过期 → 返回 null 并清除镜像', () => {
    vi.useFakeTimers();
    saveRegPending(PENDING);
    expect(readRegPending()).toEqual(PENDING);

    vi.advanceTimersByTime(5 * 60 * 1000 + 1);
    expect(readRegPending()).toBeNull();
    expect(window.sessionStorage.getItem(REG_PENDING_KEY)).toBeNull();
  });

  it('clear → 内存态与 session 镜像一并清除，之后读返回 null', () => {
    saveRegPending(PENDING);
    clearRegPending();
    expect(readRegPending()).toBeNull();
    expect(window.sessionStorage.getItem(REG_PENDING_KEY)).toBeNull();
  });

  it('空数据（未保存）→ 返回 null', () => {
    expect(readRegPending()).toBeNull();
  });
});
