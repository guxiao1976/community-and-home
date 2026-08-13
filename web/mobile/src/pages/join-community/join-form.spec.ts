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

  it('rejects building <= 0 (positive required, no hardcoded range)', () => {
    expect(validateJoinForm({ building: '0', unit: '2', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '-3', unit: '2', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '12', unit: '2', room: '301', ownership: 1 }).valid).toBe(true);
  });

  it('rejects unit <= 0 (positive required, no hardcoded range)', () => {
    expect(validateJoinForm({ building: '1', unit: '0', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '1', unit: '-1', room: '301', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '1', unit: '4', room: '301', ownership: 1 }).valid).toBe(true);
  });

  it('rejects room <= 0 (positive required, no hardcoded 3-digit range)', () => {
    expect(validateJoinForm({ building: '1', unit: '2', room: '0', ownership: 1 }).valid).toBe(false);
    expect(validateJoinForm({ building: '1', unit: '2', room: '30', ownership: 1 }).valid).toBe(true);
    expect(validateJoinForm({ building: '1', unit: '2', room: '999', ownership: 1 }).valid).toBe(true);
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
