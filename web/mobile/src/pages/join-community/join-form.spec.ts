// Unit tests for the join-community form validation + payload mapping.
import { describe, it, expect } from 'vitest';
import { OWNERSHIP_OPTIONS, validateJoinForm, joinFormToPayload } from './join-form';

describe('validateJoinForm', () => {
  it('is valid when ownership is OWNED(1) and building/unit/room are filled', () => {
    const r = validateJoinForm({ building: '1', unit: '2', room: '301', ownership: 1 });
    expect(r.valid).toBe(true);
    expect(r.errors).toEqual({});
  });

  it('is valid when ownership is RENTED(2)', () => {
    const r = validateJoinForm({ building: '3', unit: '1', room: '502', ownership: 2 });
    expect(r.valid).toBe(true);
  });

  it('requires ownership selection (null → invalid with ownership error)', () => {
    const r = validateJoinForm({ building: '1', unit: '2', room: '301', ownership: null });
    expect(r.valid).toBe(false);
    expect(r.errors.ownership).toBeTruthy();
  });

  it('rejects ownership values other than 1/2', () => {
    const r = validateJoinForm({ building: '1', unit: '2', room: '301', ownership: 0 });
    expect(r.valid).toBe(false);
    expect(r.errors.ownership).toBeTruthy();
  });

  it('validates building: positive integer, ≤ 200', () => {
    expect(validateJoinForm({ building: '0', unit: '2', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '-3', unit: '2', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '201', unit: '2', room: '301', ownership: 1 }).valid).toBe(false); // > 200
    expect(validateJoinForm({ building: '200', unit: '2', room: '301', ownership: 1 }).valid).toBe(true); // 边界
    expect(validateJoinForm({ building: '12', unit: '2', room: '301', ownership: 1 }).valid).toBe(true);
  });

  it('validates unit: positive integer, ≤ 6', () => {
    expect(validateJoinForm({ building: '1', unit: '0', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '1', unit: '-1', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '1', unit: '7', room: '301', ownership: 1 }).valid).toBe(false); // > 6
    expect(validateJoinForm({ building: '1', unit: '6', room: '301', ownership: 1 }).valid).toBe(true); // 边界
    expect(validateJoinForm({ building: '1', unit: '4', room: '301', ownership: 1 }).valid).toBe(true);
  });

  it('validates room: 3/4 位，楼层 1-55，门牌 01-04', () => {
    // 合法：502=5层02室、1102=11层02室、301=3层01室、204=2层04室、101=1层01室、5502=55层02室
    for (const room of ['502', '1102', '301', '204', '101', '5502']) {
      expect(validateJoinForm({ building: '1', unit: '2', room, ownership: 1 }).valid).toBe(true);
    }
    // 非法：位数不对 / 门牌非01-04（如14、05、00、99）/ 楼层>55（如56）
    for (const room of ['0', '30', '999', '1205', '552', '5602', '5514']) {
      expect(validateJoinForm({ building: '1', unit: '2', room, ownership: 1 }).valid).toBe(false);
    }
  });

  it('reports all errors at once when form is empty', () => {
    const r = validateJoinForm({ building: '', unit: '', room: '', ownership: null });
    expect(r.valid).toBe(false);
    expect(r.errors.ownership).toBeTruthy();
    expect(r.errors.building).toBeTruthy();
    expect(r.errors.unit).toBeTruthy();
    expect(r.errors.room).toBeTruthy();
  });
});

describe('joinFormToPayload', () => {
  it('converts string inputs to numbers and keeps ownership', () => {
    expect(joinFormToPayload({ building: '3', unit: '1', room: '502', ownership: 2 })).toEqual({
      building: 3,
      unit: 1,
      room: 502,
      ownership: 2,
    });
  });
});

describe('OWNERSHIP_OPTIONS', () => {
  it('exposes 自有(OWNED=1) and 租住(RENTED=2)', () => {
    expect(OWNERSHIP_OPTIONS).toEqual([
      { value: 1, label: '自有' },
      { value: 2, label: '租住' },
    ]);
  });
});
