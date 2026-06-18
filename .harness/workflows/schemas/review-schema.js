// ============================================================
// Review Schema — 代码审查结果结构
// ============================================================

export const REVIEW_SCHEMA = {
  type: "object",
  properties: {
    verdict: {
      type: "string",
      enum: ["PASS", "FAIL"],
      description: "最终判定"
    },
    lens: {
      type: "string",
      description: "审查视角"
    },
    summary: {
      type: "string",
      description: "审查摘要"
    },
    criticals: {
      type: "array",
      items: {
        type: "object",
        properties: {
          file: { type: "string" },
          line: { type: "number" },
          issue: { type: "string" },
          suggestion: { type: "string" }
        },
        required: ["file", "issue"]
      }
    },
    warnings: {
      type: "array",
      items: {
        type: "object",
        properties: {
          file: { type: "string" },
          issue: { type: "string" }
        }
      }
    }
  },
  required: ["verdict", "lens", "summary", "criticals"]
}
