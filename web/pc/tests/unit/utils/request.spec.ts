// lossless-json bigint precision tests for Snowflake ID handling
import { describe, it, expect } from 'vitest'
import { parse, stringify, isLosslessNumber, LosslessNumber } from 'lossless-json'

describe('lossless-json 大整数解析', () => {
  // F1-01: 19位ID精确解析
  it('应精确解析 19 位雪花 ID', () => {
    const json = '{"id": 1234567890123456789}'
    const result = parse(json) as { id: unknown }
    // lossless-json 将大数字存为 LosslessNumber，转 string 时保持精度
    expect(String(result.id)).toBe('1234567890123456789')
    expect(isLosslessNumber(result.id)).toBe(true)
  })

  // F1-02: 安全范围内整数正常解析
  it('应正常解析安全范围内的整数', () => {
    const json = '{"id": 123456, "status": 1}'
    const result = parse(json) as { id: unknown; status: unknown }
    expect(isLosslessNumber(result.id)).toBe(true)
    expect(isLosslessNumber(result.status)).toBe(true)
    // LosslessNumber.valueOf() returns JS number for safe values
    expect((result.id as LosslessNumber).valueOf()).toBe(123456)
    expect((result.status as LosslessNumber).valueOf()).toBe(1)
  })

  // F1-03: 17位ID精确解析（边界测试 — 超过安全范围）
  it('应精确解析超过 Number.MAX_SAFE_INTEGER 的数值', () => {
    const unsafeInteger = Number.MAX_SAFE_INTEGER + 1
    const json = `{"id": ${unsafeInteger}}`
    const result = parse(json) as { id: unknown }
    expect(String(result.id)).toBe('9007199254740992')
  })

  // F1-04: 嵌套对象中的大整数
  it('应精确解析嵌套对象中的大整数', () => {
    const json = '{"user": {"id": 1234567890123456789}, "roles": [{"id": 9876543210987654321}]}'
    const result = parse(json) as { user: { id: unknown }; roles: Array<{ id: unknown }> }
    expect(String(result.user.id)).toBe('1234567890123456789')
    expect(String(result.roles[0].id)).toBe('9876543210987654321')
  })

  // F1-05: 数组中的大整数
  it('应保留 JSON 数组中所有大整数的精度', () => {
    const json = '[{"id": 1111111111111111111}, {"id": 2222222222222222222}]'
    const result = parse(json) as Array<{ id: unknown }>
    expect(String(result[0].id)).toBe('1111111111111111111')
    expect(String(result[1].id)).toBe('2222222222222222222')
  })

  // F1-06: null 和空字符串不受影响
  it('应正确处理 null 和空字符串', () => {
    const json = '{"id": null, "name": ""}'
    const result = parse(json) as { id: unknown; name: unknown }
    expect(result.id).toBeNull()
    expect(result.name).toBe('')
  })

  // F1-07: 序列化回 JSON 保持精度
  it('序列化回 JSON 应保持大整数精度', () => {
    const original = '{"id": 1234567890123456789}'
    const parsed = parse(original)
    const serialized = stringify(parsed) as string
    // 重新解析验证
    const reparsed = parse(serialized) as { id: unknown }
    expect(String(reparsed.id)).toBe('1234567890123456789')
  })
})

describe('reviveLargeNumbers — 安全整数保持为 number，大整数转为 string', () => {
  // Custom reviver: keep safe integers as numbers, convert large integers to strings.
  function reviveLargeNumbers(_key: string, value: unknown): unknown {
    if (isLosslessNumber(value)) {
      const asNumber = Number((value as LosslessNumber).toString())
      if (Number.isSafeInteger(asNumber)) {
        return asNumber
      }
      return (value as LosslessNumber).toString()
    }
    return value
  }

  // R1-01: 安全范围内的整数返回 number
  it('安全范围内的 id 应保持为 number', () => {
    const json = '{"id": 123456, "status": 1}'
    const result = parse(json, reviveLargeNumbers) as { id: unknown; status: unknown }
    expect(typeof result.id).toBe('number')
    expect(result.id).toBe(123456)
    expect(typeof result.status).toBe('number')
    expect(result.status).toBe(1)
  })

  // R1-02: 19位雪花 ID 应转为 string
  it('19 位雪花 ID 应转为 string', () => {
    const json = '{"id": 1234567890123456789}'
    const result = parse(json, reviveLargeNumbers) as { id: unknown }
    expect(typeof result.id).toBe('string')
    expect(result.id).toBe('1234567890123456789')
  })

  // R1-03: 嵌套对象中的大 ID 转为 string
  it('嵌套对象中的大整数转为 string，小整数保持 number', () => {
    const json = '{"user": {"id": 1234567890123456789, "status": 1}, "total": 100}'
    const result = parse(json, reviveLargeNumbers) as {
      user: { id: unknown; status: unknown }
      total: unknown
    }
    expect(typeof result.user.id).toBe('string')
    expect(result.user.id).toBe('1234567890123456789')
    expect(typeof result.user.status).toBe('number')
    expect(result.user.status).toBe(1)
    expect(typeof result.total).toBe('number')
    expect(result.total).toBe(100)
  })

  // R1-04: 已经为 JSON string 的 ID 保持为 string
  it('后端返回的字符串 ID 应保持为 string', () => {
    // 模拟后端 json:",string" 的输出
    const json = '{"id": "1234567890123456789", "status": 1}'
    const result = parse(json, reviveLargeNumbers) as { id: unknown; status: unknown }
    expect(typeof result.id).toBe('string')
    expect(result.id).toBe('1234567890123456789')
    expect(typeof result.status).toBe('number')
  })

  // R1-05: 数组中的混合大小 ID
  it('数组中的混合大小整数应正确分类', () => {
    const json = '[{"id": 1}, {"id": 1234567890123456789}, {"id": 999}]'
    const result = parse(json, reviveLargeNumbers) as Array<{ id: unknown }>
    expect(typeof result[0].id).toBe('number')
    expect(result[0].id).toBe(1)
    expect(typeof result[1].id).toBe('string')
    expect(result[1].id).toBe('1234567890123456789')
    expect(typeof result[2].id).toBe('number')
    expect(result[2].id).toBe(999)
  })
})
