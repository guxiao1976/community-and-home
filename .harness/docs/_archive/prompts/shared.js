// ============================================================
// Shared Context — 共享状态管理
// ============================================================
//
// 所有 Prompt 模块从这里获取共享上下文（SVC_DIR, args 等）

let sharedContext = {}

export function setSharedContext(ctx) {
  sharedContext = ctx
}

export function getContext() {
  return sharedContext
}

// 便捷访问器
export function getSvcDir() {
  return sharedContext.SVC_DIR || ''
}

export function getArgs() {
  return sharedContext.args || {}
}

export function getServiceName() {
  return sharedContext.args?.serviceName || ''
}

export function isFrontend() {
  return (sharedContext.SVC_DIR || '').startsWith('web/')
}
