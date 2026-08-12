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

// 楼号/单元号/房号区间与 user.proto JoinCommunityRequest 注释对齐（building 1-150 / unit 1-5 / room 3位数字）。
export function validateJoinForm(form: JoinFormState): JoinFormResult {
  const errors: JoinFormErrors = {};

  if (form.ownership !== 1 && form.ownership !== 2) {
    errors.ownership = '请选择房屋权属（自有/租住）';
  }

  const building = Number(form.building);
  if (!String(form.building).trim() || !Number.isInteger(building) || building < 1 || building > 150) {
    errors.building = '请输入楼号（1-150）';
  }

  const unit = Number(form.unit);
  if (!String(form.unit).trim() || !Number.isInteger(unit) || unit < 1 || unit > 5) {
    errors.unit = '请输入单元号（1-5）';
  }

  const room = Number(form.room);
  if (!String(form.room).trim() || !Number.isInteger(room) || room < 100 || room > 999) {
    errors.room = '请输入3位房号（如301）';
  }

  return { valid: Object.keys(errors).length === 0, errors };
}

// 前端表单（字符串输入）→ joinCommunity 载荷（int 楼/单元/房号 + 权属枚举值）。
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
