// pending-join.ts — 加入小区流程跨页一次性「待加入小区」数据的唯一契约源
//
// 背景：join-community.vue 选中小区的 4 步向导原在页内弹「自有/租住 + 楼单元房号」modal。
// 新设计：选中小区后不再弹窗，改为把选中小区 {id, name, address} 存入本模块 → 新页
// join-choice（2 个身份选项）分流：
//   - 业主路径 → join-residence 填房号（joinCommunity + applyRole owner）
//   - 其他身份认证路径 → 我的页申请角色（communityStore.pendingCommunityId）
//
// 载体与 reg-pending 一致：模块级内存变量为主载体（页面栈内导航即可传递，不落持久化）；
// 仅 H5 镜像到 sessionStorage（{data, expiresAt}，TTL 30 分钟，非敏感数据可放宽），绝不
// localStorage；非 H5 环境统一走内存态。sessionStorage 访问一律 try/catch 容错。
//
// SEE: [[frontend-cross-page-storage-contract]] — 跨页共享 key/结构收敛到单一共享模块，禁止两端各写 magic string

/** 跨页共享的临时键（唯一契约源） */
export const PENDING_JOIN_KEY = 'pending_join';

/** 跨页暂存的待加入小区 */
export interface PendingJoin {
  communityId: string;
  communityName: string;
  address?: string;
}

/** sessionStorage 镜像的 TTL（30 分钟；非一次性验证码等敏感数据，时长可放宽） */
const TTL_MS = 30 * 60 * 1000;

/** sessionStorage 镜像载荷：data + 过期时间戳 */
interface PendingJoinEnvelope {
  data: PendingJoin;
  expiresAt: number;
}

/** 模块级内存态主载体（页面栈内导航直接可读） */
let memory: PendingJoinEnvelope | null = null;

/** H5 环境才有 window；非 H5（小程序等）统一走内存态 */
const isH5 = (): boolean => typeof window !== 'undefined';

/** 由 envelope 判定有效数据：过期即整体清除并返回 null */
function validFrom(envelope: PendingJoinEnvelope | null): PendingJoin | null {
  if (!envelope) return null;
  if (envelope.expiresAt <= Date.now()) {
    clearPendingJoin();
    return null;
  }
  return envelope.data;
}

/** 保存跨页待加入小区：内存态 + H5 sessionStorage 镜像（先清旧再写） */
export function savePendingJoin(data: PendingJoin): void {
  const envelope: PendingJoinEnvelope = { data, expiresAt: Date.now() + TTL_MS };
  memory = envelope;
  if (isH5()) {
    try {
      window.sessionStorage.removeItem(PENDING_JOIN_KEY);
      window.sessionStorage.setItem(PENDING_JOIN_KEY, JSON.stringify(envelope));
    } catch {
      // sessionStorage 不可用（隐私模式/被禁用）→ 降级内存态，不抛出
    }
  }
}

/** 读取跨页待加入小区：内存态优先，H5 刷新导致内存丢失时回退 session 镜像 */
export function readPendingJoin(): PendingJoin | null {
  const fromMemory = validFrom(memory);
  if (fromMemory) return fromMemory;
  if (isH5()) {
    try {
      const raw = window.sessionStorage.getItem(PENDING_JOIN_KEY);
      if (raw) {
        const fromSession = validFrom(JSON.parse(raw) as PendingJoinEnvelope);
        if (fromSession) {
          memory = { data: fromSession, expiresAt: Date.now() + TTL_MS };
          return fromSession;
        }
        window.sessionStorage.removeItem(PENDING_JOIN_KEY);
      }
    } catch {
      // 解析/访问失败 → 按无数据处理
    }
  }
  return null;
}

/** 清除跨页待加入小区：内存态 + session 镜像一并清 */
export function clearPendingJoin(): void {
  memory = null;
  if (isH5()) {
    try {
      window.sessionStorage.removeItem(PENDING_JOIN_KEY);
    } catch {
      // 忽略存储错误
    }
  }
}
