// Identity Service API

import request from '@/utils/request';
import type {
  LoginRequest,
  LoginSmsRequest,
  RegisterRequest,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  User,
  UserFilter,
  Role,
  Permission,
  HomeownerVerification
} from '@common/types/identity';
import type { PaginatedResponse, PaginationParams } from '@common/types/common';
import { getPublicKey, encryptWithPublicKey } from '@/utils/crypto';
import { getDeviceId, getDeviceType } from '@/utils/device';

/**
 * Fetch RSA public key from auth service.
 * GET /api/auth/public-key → { public_key: string }
 */
export { getPublicKey } from '@/utils/crypto';

/**
 * Login with phone and password (RSA encrypted).
 */
export async function login(phone: string, password: string): Promise<LoginResponse> {
  const [publicKey, deviceId] = await Promise.all([
    getPublicKey(),
    getDeviceId()
  ]);

  const [encryptedPhone, encryptedPassword] = await Promise.all([
    encryptWithPublicKey(phone, publicKey),
    encryptWithPublicKey(password, publicKey)
  ]);

  const data: LoginRequest = {
    encryptedPhone,
    encryptedPassword,
    deviceId,
    deviceType: getDeviceType()
  };

  return request.post<LoginResponse>('/api/auth/login', data);
}

/**
 * Login with phone and SMS code (phone RSA encrypted).
 */
export async function loginWithSms(phone: string, smsCode: string): Promise<LoginResponse> {
  const [publicKey, deviceId] = await Promise.all([
    getPublicKey(),
    getDeviceId()
  ]);

  const encryptedPhone = await encryptWithPublicKey(phone, publicKey);

  const data: LoginSmsRequest = {
    encryptedPhone,
    smsCode,
    deviceId,
    deviceType: getDeviceType()
  };

  return request.post<LoginResponse>('/api/auth/login/sms', data);
}

/**
 * Register new user (phone + password RSA encrypted).
 */
export async function register(registerData: {
  phone: string;
  password?: string;
  smsCode: string;
  nickname: string;
}): Promise<LoginResponse> {
  const [publicKey, deviceId] = await Promise.all([
    getPublicKey(),
    getDeviceId()
  ]);

  const encryptedPhone = await encryptWithPublicKey(registerData.phone, publicKey);
  const encryptedPassword = registerData.password
    ? await encryptWithPublicKey(registerData.password, publicKey)
    : undefined;

  const data: RegisterRequest = {
    encryptedPhone,
    encryptedPassword,
    smsCode: registerData.smsCode,
    nickname: registerData.nickname,
    deviceId,
    deviceType: getDeviceType()
  };

  return request.post<LoginResponse>('/api/auth/register', data);
}

/**
 * Send SMS verification code
 */
export function sendSms(phone: string) {
  return request.post<null>('/api/auth/sms/send', { phone });
}

/**
 * Refresh access token
 */
export function refreshToken(data: RefreshTokenRequest) {
  return request.post<RefreshTokenResponse>('/api/auth/token/refresh', data);
}

/**
 * Logout
 */
export function logout() {
  const deviceId = getDeviceId();
  return request.post<null>('/api/auth/logout', { deviceId });
}

/**
 * Get users list with pagination and filters
 */
export async function getUsers(params?: UserFilter & PaginationParams) {
  const data = await request.get<{ users: User[]; page: { total: number } }>('/api/users', { params });
  return { list: data.users, total: data.page.total };
}

/**
 * Get user by ID
 */
export async function getUserById(id: string) {
  const data = await request.get<{ user: User }>(`/api/users/${id}`);
  return data.user;
}

/**
 * Create new user
 */
export function createUser(data: {
  phone: string;
  password: string;
  nickname?: string;
  user_type: number;
  scope?: string;
}) {
  return request.post<{ id: string }>('/api/users', data);
}

/**
 * Update user
 */
export function updateUser(id: string, data: {
  nickname?: string;
  scope?: string;
}) {
  return request.put<null>(`/api/users/${id}`, data);
}

/**
 * Disable user account (set status to 2)
 */
export function disableUser(id: string) {
  return request.put<null>(`/api/users/${id}`, { status: 2 });
}

/**
 * Enable user account (set status to 1)
 */
export function enableUser(id: string) {
  return request.put<null>(`/api/users/${id}`, { status: 1 });
}

/**
 * Get user permissions
 */
export function getUserPermissions(userId: string) {
  return request.get<{ permissionCodes: string[] }>(`/api/perm/users/${userId}/permissions`);
}

/**
 * Get user roles
 */
export function getUserRoles(userId: string) {
  return request.get<{ roles: UserRole[] }>(`/api/perm/users/${userId}/roles`);
}

/**
 * Get users assigned to a role
 */
export function getRoleUsers(roleId: string, page = 1, pageSize = 20) {
  return request.get<{
    users: Array<{ userId: string; phone: string; nickname: string }>;
    page: number; pageSize: number; total: number; totalPages: number;
  }>(`/api/perm/roles/${roleId}/users`, { params: { page, pageSize } });
}

/**
 * Assign role to user with scope (admin operation).
 * Matches backend AssignUserRoleReq: { userId, roleId, scopeType, scopeId }
 */
export function assignUserRole(data: {
  userId: string;
  roleId: string;
  scopeType: string;
  scopeId: string;
}) {
  return request.post<null>('/api/perm/user-roles', {
    userId: data.userId,
    roleId: data.roleId,
    scopeType: data.scopeType,
    scopeId: data.scopeId,
  });
}

/**
 * Revoke role from user (admin operation).
 * Matches backend RevokeUserRoleReq: { userId, roleId, scopeType?, scopeId? }
 */
export function revokeUserRole(data: {
  userId: string;
  roleId: string;
  scopeType?: string;
  scopeId?: string;
}) {
  return request.delete<null>('/api/perm/user-roles', {
    data: {
      userId: data.userId,
      roleId: data.roleId,
      scopeType: data.scopeType,
      scopeId: data.scopeId,
    },
  });
}

/**
 * Get all permissions
 */
export function getPermissions() {
  return request.get<{ permissions: Permission[] }>('/api/perm/permissions');
}

/**
 * Auto-discover: 扫描已注册路由，自动注册缺失的 API 权限
 */
export function autoDiscoverPermissions() {
  return request.post<{
    added: Array<{ id: string; parentId: string; name: string; code: string; path: string }>;
    total: number;
    message: string;
  }>('/api/perm/permissions/auto-discover');
}

/**
 * Get all roles
 */
export function getRoles(params?: PaginationParams) {
  return request.get<PaginatedResponse<Role>>('/api/perm/roles', { params });
}

/**
 * Get role by ID
 */
export function getRoleById(id: string) {
  return request.get<{ role: Role }>(`/api/perm/roles/${id}`);
}

/**
 * Create role
 */
export function createRole(data: {
  name: string;
  code: string;
  description?: string;
}) {
  return request.post<{ id: string }>('/api/perm/roles', data);
}

/**
 * Update role
 */
export function updateRole(id: string, data: {
  name?: string;
  description?: string;
}) {
  return request.put<null>(`/api/perm/roles/${id}`, data);
}

/**
 * Delete role
 */
export function deleteRole(id: string) {
  return request.delete<null>(`/api/perm/roles/${id}`);
}

/**
 * Get role permissions
 */
export function getRolePermissions(roleId: string) {
  return request.get<{ permissionIds: string[] }>(`/api/perm/roles/${roleId}/permissions`);
}

/**
 * Assign permissions to role
 */
export function assignRolePermissions(roleId: string, permissionIds: string[]) {
  return request.post<null>(`/api/perm/roles/${roleId}/permissions`, {
    permissionIds: permissionIds
  });
}

/**
 * Get homeowner verifications list
 */
export function getVerifications(params?: {
  status?: number;
  start_date?: string;
  end_date?: string;
} & PaginationParams) {
  return request.get<PaginatedResponse<HomeownerVerification>>('/api/verifications', { params });
}

/**
 * Get verification by ID
 */
export function getVerificationById(id: string) {
  return request.get<HomeownerVerification>(`/api/verifications/${id}`);
}

/**
 * Review homeowner verification
 */
export function reviewVerification(id: string, data: {
  status: number;
  review_notes?: string;
}) {
  return request.post<null>(`/api/verifications/${id}/review`, data);
}
