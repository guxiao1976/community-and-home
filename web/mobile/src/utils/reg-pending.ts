// reg-pending.ts — 登录协议流程跨页一次性注册数据的唯一契约源
//
// 背景：login.vue（50001 未注册）暂存 {phone, smsCode, deviceId, nickname}，
// 跳协议页 agreement.vue 确认注册后消费。此前两端各自手写 'reg_pending' magic
// string + uni.setStorageSync 持久化到 H5 localStorage，存在：
//   1. smsCode 一次性验证码被持久化残留，共享设备可被复用
//   2. 两端 magic string 各写一份，一端改名另一端静默失效
//   3. 跨页一次性敏感数据应优先内存态载体，而非持久化
//
// 本模块作为唯一契约源收敛 key 与 RegPending 结构：
//   - 模块级内存变量为主载体：页面栈内导航即可传递，不落持久化
//   - 仅 H5 镜像到 sessionStorage（{data, expiresAt}，TTL 5 分钟），绝不 localStorage
//   - 非 H5 环境统一走内存态（本改动目标是消灭 localStorage 持久化，跨端后续再议）
//   - sessionStorage 访问一律 try/catch 容错（隐私模式 / 被禁用时降级内存态）
//
// SEE: [[sms-code-persist-localstorage]] — 一次性验证码禁止落 localStorage 持久化残留
// SEE: [[frontend-cross-page-storage-contract]] — 跨页共享 key/结构收敛到单一共享模块，禁止两端各写 magic string
// SEE: [[cross-page-sensitive-temp-data-storage]] — 跨页一次性敏感数据优先内存态载体

/** 跨页共享的临时键（唯一契约源） */
export const REG_PENDING_KEY = 'reg_pending';

/** 跨页暂存的一键注册所需数据 */
export interface RegPending {
  phone: string;
  smsCode: string;
  deviceId: string;
  nickname: string;
}

/** sessionStorage 镜像的 TTL（5 分钟） */
const TTL_MS = 5 * 60 * 1000;

/** sessionStorage 镜像载荷：data + 过期时间戳 */
interface RegPendingEnvelope {
  data: RegPending;
  expiresAt: number;
}

/** 模块级内存态主载体（页面栈内导航直接可读） */
let memory: RegPendingEnvelope | null = null;

/** H5 环境才有 window；非 H5（小程序等）统一走内存态 */
const isH5 = (): boolean => typeof window !== 'undefined';

/** 由 envelope 判定有效数据：过期即整体清除并返回 null */
function validFrom(envelope: RegPendingEnvelope | null): RegPending | null {
  if (!envelope) return null;
  if (envelope.expiresAt <= Date.now()) {
    clearRegPending();
    return null;
  }
  return envelope.data;
}

/** 保存跨页一次性注册数据：内存态 + H5 sessionStorage 镜像（先清旧再写） */
export function saveRegPending(data: RegPending): void {
  const envelope: RegPendingEnvelope = { data, expiresAt: Date.now() + TTL_MS };
  memory = envelope;
  if (isH5()) {
    try {
      window.sessionStorage.removeItem(REG_PENDING_KEY);
      window.sessionStorage.setItem(REG_PENDING_KEY, JSON.stringify(envelope));
    } catch {
      // sessionStorage 不可用（隐私模式/被禁用）→ 降级内存态，不抛出
    }
  }
}

/** 读取跨页一次性注册数据：内存态优先，H5 刷新导致内存丢失时回退 session 镜像 */
export function readRegPending(): RegPending | null {
  const fromMemory = validFrom(memory);
  if (fromMemory) return fromMemory;
  if (isH5()) {
    try {
      const raw = window.sessionStorage.getItem(REG_PENDING_KEY);
      if (raw) {
        const fromSession = validFrom(JSON.parse(raw) as RegPendingEnvelope);
        if (fromSession) {
          memory = { data: fromSession, expiresAt: Date.now() + TTL_MS };
          return fromSession;
        }
        window.sessionStorage.removeItem(REG_PENDING_KEY);
      }
    } catch {
      // 解析/访问失败 → 按无数据处理
    }
  }
  return null;
}

/** 清除跨页一次性注册数据：内存态 + session 镜像一并清 */
export function clearRegPending(): void {
  memory = null;
  if (isH5()) {
    try {
      window.sessionStorage.removeItem(REG_PENDING_KEY);
    } catch {
      // 忽略存储错误
    }
  }
}
