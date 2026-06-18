// Form validation rules

import type { FormItemRule } from 'element-plus';

/**
 * Phone number validation rule (Chinese mobile format)
 */
export const phoneRule: FormItemRule = {
  pattern: /^1[3-9]\d{9}$/,
  message: '请输入正确的手机号码',
  trigger: 'blur'
};

/**
 * Required phone number rule
 */
export const requiredPhoneRule: FormItemRule[] = [
  { required: true, message: '请输入手机号码', trigger: 'blur' },
  phoneRule
];

/**
 * Password validation rule (min 8 characters, with uppercase, lowercase, number, special char)
 */
export const passwordRule: FormItemRule = {
  pattern: /^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]{8,}$/,
  message: '密码至少8位，包含大小写字母、数字和特殊字符',
  trigger: 'blur'
};

/**
 * Required password rule
 */
export const requiredPasswordRule: FormItemRule[] = [
  { required: true, message: '请输入密码', trigger: 'blur' },
  { min: 8, message: '密码至少8位', trigger: 'blur' },
  passwordRule
];

/**
 * ID card validation rule (Chinese ID card format)
 */
export const idCardRule: FormItemRule = {
  pattern: /^\d{17}[\dXx]$/,
  message: '请输入正确的身份证号码',
  trigger: 'blur'
};

/**
 * Required ID card rule
 */
export const requiredIdCardRule: FormItemRule[] = [
  { required: true, message: '请输入身份证号码', trigger: 'blur' },
  idCardRule
];

/**
 * Email validation rule
 */
export const emailRule: FormItemRule = {
  type: 'email',
  message: '请输入正确的邮箱地址',
  trigger: 'blur'
};

/**
 * URL validation rule
 */
export const urlRule: FormItemRule = {
  type: 'url',
  message: '请输入正确的URL地址',
  trigger: 'blur'
};

/**
 * Nickname validation rule (2-20 characters)
 */
export const nicknameRule: FormItemRule[] = [
  { required: true, message: '请输入昵称', trigger: 'blur' },
  { min: 2, max: 20, message: '昵称长度为2-20个字符', trigger: 'blur' }
];

/**
 * SMS code validation rule (6 digits)
 */
export const smsCodeRule: FormItemRule[] = [
  { required: true, message: '请输入验证码', trigger: 'blur' },
  { pattern: /^\d{6}$/, message: '验证码为6位数字', trigger: 'blur' }
];

/**
 * Required field rule
 */
export const requiredRule = (message: string): FormItemRule => ({
  required: true,
  message,
  trigger: 'blur'
});

/**
 * Number range rule
 */
export const numberRangeRule = (min: number, max: number, message?: string): FormItemRule => ({
  type: 'number',
  min,
  max,
  message: message || `请输入${min}-${max}之间的数字`,
  trigger: 'blur'
});

/**
 * String length rule
 */
export const lengthRule = (min: number, max: number, message?: string): FormItemRule => ({
  min,
  max,
  message: message || `长度为${min}-${max}个字符`,
  trigger: 'blur'
});
