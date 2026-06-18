// Error code constants
export const ErrorCode = {
  Success: 0,
  InvalidParameter: 400,
  Unauthorized: 401,
  Forbidden: 403,
  NotFound: 404,
  InternalServerError: 500,
  DatabaseError: 501,
  CacheError: 502,
  RPCError: 503
} as const;

export type ErrorCode = typeof ErrorCode[keyof typeof ErrorCode];

export const ERROR_MESSAGES: Record<number, string> = {
  [ErrorCode.Success]: '操作成功',
  [ErrorCode.InvalidParameter]: '参数错误',
  [ErrorCode.Unauthorized]: '未授权',
  [ErrorCode.Forbidden]: '权限不足',
  [ErrorCode.NotFound]: '资源不存在',
  [ErrorCode.InternalServerError]: '服务器内部错误',
  [ErrorCode.DatabaseError]: '数据库错误',
  [ErrorCode.CacheError]: '缓存错误',
  [ErrorCode.RPCError]: 'RPC服务错误'
};
