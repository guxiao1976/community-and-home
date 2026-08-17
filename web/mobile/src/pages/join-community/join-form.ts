// 加入小区表单：房屋权属 + 楼/单元/房号校验与载荷映射。
// 权属取值与 user.proto `CommunityOwnership` 对齐：OWNED=1 / RENTED=2，前端仅传 1 或 2（0 为非法值，后端 10040 拒绝）。
// SEE: [[testing-discipline]] — 校验函数按正常/空值/边界/错误路径全覆盖，配套 table-driven 测试（join-form.spec.ts）。
export const OWNERSHIP_OPTIONS = [
  { value: 1, label: '自有' },
  { value: 2, label: '租住' },
] as const;

export interface JoinFormState {
  // building/unit/room 允许 string | number：Vue3 v-model 对 `<input type="number">`
  // 会自动把值转成 number，统一在 validateJoinForm 内 String() 归一化处理。
  building: string | number;
  unit: string | number;
  room: string | number;
  ownership: number | null;
}

export interface JoinFormErrors {
  ownership?: string;
  building?: string;
  unit?: string;
  room?: string;
}

export interface JoinFormResult {
  valid: boolean;
  errors: JoinFormErrors;
}

// 楼/单元/房号业务规则（用户确认）：
//   - 楼号：正整数，≤ 200
//   - 单元号：正整数，≤ 6
//   - 房号：3 或 4 位数字；楼号=除后 2 位外的前 1-2 位（1-55 层），门牌号=后 2 位（01-04）
//     例：502 = 5层02室；1102 = 11层02室。正则：`^([1-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-5])(0[1-4])$`
// SEE: [[frontend-business-rule-hardcode]] — 前端做 UX 即时校验；后端 JoinCommunity 仍须权威校验（防绕过）
export const ROOM_REGEX = /^([1-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-5])(0[1-4])$/;
export const MAX_BUILDING = 200;
export const MAX_UNIT = 6;

export function validateJoinForm(form: JoinFormState): JoinFormResult {
  const errors: JoinFormErrors = {};

  if (form.ownership !== 1 && form.ownership !== 2) {
    errors.ownership = '请选择房屋权属（自有/租住）';
  }

  const building = Number(form.building);
  if (!String(form.building).trim() || !Number.isInteger(building) || building < 1) {
    errors.building = '请输入楼号';
  } else if (building > MAX_BUILDING) {
    errors.building = `楼号不能超过 ${MAX_BUILDING}`;
  }

  const unit = Number(form.unit);
  if (!String(form.unit).trim() || !Number.isInteger(unit) || unit < 1) {
    errors.unit = '请输入单元号';
  } else if (unit > MAX_UNIT) {
    errors.unit = `单元号不能超过 ${MAX_UNIT}`;
  }

  const room = String(form.room).trim();
  if (!room || !/^\d{3,4}$/.test(room)) {
    errors.room = '请输入房号（3或4位数字）';
  } else if (!ROOM_REGEX.test(room)) {
    errors.room = '房号格式有误：楼层≤55，门牌号01-04（如 502 或 1102）';
  }

  return { valid: Object.keys(errors).length === 0, errors };
}

// 前端表单（字符串输入）→ 归一化（int 楼/单元/房号 + 权属枚举值）。
// 新模型下 join-residence 使用：building/unit/room 转 string 进 bindResidence，ownership 决定 applyRole 的 role_code。
export function joinFormToPayload(form: JoinFormState): {
  building: number;
  unit: number;
  room: number;
  ownership: number;
} {
  return {
    building: Number(form.building),
    unit: Number(form.unit),
    room: Number(form.room),
    ownership: form.ownership as number,
  };
}
