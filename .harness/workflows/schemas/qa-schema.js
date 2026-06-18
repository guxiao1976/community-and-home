// ============================================================
// QA Schema — 质量检查结果结构
// ============================================================

export const QA_SCHEMA = {
  type: "object",
  properties: {
    verdict: {
      type: "string",
      enum: ["PASS", "FAIL"],
      description: "最终判定结果"
    },
    summary: {
      type: "string",
      description: "检查摘要（简短）"
    },
    checks: {
      type: "array",
      items: {
        type: "object",
        properties: {
          name: { type: "string" },
          status: { type: "string", enum: ["PASS", "FAIL", "WARN"] },
          detail: { type: "string" }
        },
        required: ["name", "status"]
      }
    },
    failures: {
      type: "array",
      items: {
        type: "object",
        properties: {
          check: { type: "string" },
          reason: { type: "string" },
          file: { type: "string" }
        },
        required: ["check", "reason"]
      }
    }
  },
  required: ["verdict", "summary", "checks"]
}
