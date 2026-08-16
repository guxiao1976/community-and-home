// Unit tests — community.ts image-attachment whitelist helper.
// isImageAttachment is a branch (predicate) function → TDD RED→GREEN.
// SEE: [[frontend-business-rule-hardcode]] — 白名单须与 services/file-service guard/magic.go 对齐
// SEE: [[snake-camel-field-mismatch]] — wire file_type 为 file-service 嗅探落库的规范小写扩展名
import { describe, it, expect } from 'vitest';
import { isImageAttachment } from '@/api/community';

describe('isImageAttachment — file_type 图片白名单分发（REQ-NDP-2）', () => {
  it('白名单小写扩展名 png/jpg/jpeg/gif → true', () => {
    expect(isImageAttachment('png')).toBe(true);
    expect(isImageAttachment('jpg')).toBe(true);
    expect(isImageAttachment('jpeg')).toBe(true);
    expect(isImageAttachment('gif')).toBe(true);
  });

  it('文档扩展名 pdf/doc/docx → false（走文档分支 REQ-NDP-3）', () => {
    expect(isImageAttachment('pdf')).toBe(false);
    expect(isImageAttachment('doc')).toBe(false);
    expect(isImageAttachment('docx')).toBe(false);
  });

  it('缺失/无法识别 file_type（undefined/空串/未知）→ false 不崩溃（REQ-NDP-4 场景 2）', () => {
    expect(isImageAttachment(undefined)).toBe(false);
    expect(isImageAttachment('')).toBe(false);
    expect(isImageAttachment('application/octet-stream')).toBe(false);
  });
});
