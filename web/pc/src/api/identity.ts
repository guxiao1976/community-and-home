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

/**
 * Login with phone and password
 */
export function login(data: LoginRequest) {
  return request.post<LoginResponse>('/api/identity/auth/login', data);
}

/**
 * Login with phone and SMS code
 */
export function loginWithSms(data: LoginSmsRequest) {
  return request.post<LoginResponse>('/api/identity/auth/login/sms', data);
}

/**
 * Register new user
 */
export function register(data: RegisterRequest) {
  return request.post<LoginResponse>('/api/identity/auth/register', data);
}

/**
 * Send SMS verification code
 */
export function sendSms(phone: string) {
  return request.post<null>('/api/identity/auth/sms/send', { phone });
}

/**
 * Refresh access token
 */
export function refreshToken(data: RefreshTokenRequest) {
  return request.post<RefreshTokenResponse>('/api/identity/auth/token/refresh', data);
}

/**
 * Logout
 */
export function logout() {
  return request.post<null>('/api/identity/auth/logout');
}

/**
 * Get users list with pagination and filters
 */
export function getUsers(params?: UserFilter & PaginationParams) {
  return request.get<PaginatedResponse<User>>('/api/identity/users', { params });
}

/**
 * Get user by ID
 */
export function getUserById(id: number) {
  return request.get<User>(`/api/identity/users/${id}`);
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
  return request.post<{ id: number }>('/api/identity/users', data);
}

/**
 * Update user
 */
export function updateUser(id: number, data: {
  nickname?: string;
  scope?: string;
}) {
  return request.put<null>(`/api/identity/users/${id}`, data);
}

/**
 * Disable user account
 */
export function disableUser(id: number) {
  return request.post<null>(`/api/identity/users/${id}/disable`);
}

/**
 * Enable user account
 */
export function enableUser(id: number) {
  return request.post<null>(`/api/identity/users/${id}/enable`);
}

/**
 * Get user permissions
 */
export function getUserPermissions(userId: number) {
  return request.get<{ permissions: string[]; menus: Permission[] }>(`/api/identity/users/${userId}/permissions`);
}

/**
 * Get user roles
 */
export function getUserRoles(userId: number) {
  return request.get<Role[]>(`/api/identity/users/${userId}/roles`);
}

/**
 * Assign role to user
 */
export function assignUserRole(userId: number, roleId: number) {
  return request.post<null>(`/api/identity/users/${userId}/roles`, { role_id: roleId });
}

/**
 * Remove role from user
 */
export function removeUserRole(userId: number, roleId: number) {
  return request.delete<null>(`/api/identity/users/${userId}/roles/${roleId}`);
}

/**
 * Get all roles (no pagination)
 */
export function getAllRoles() {
  return request.get<Role[]>('/api/identity/roles/all');
}

/**
 * Get all permissions
 */
export function getPermissions() {
  return request.get<Permission[]>('/api/identity/permissions');
}

/**
 * Get all roles
 */
export function getRoles(params?: PaginationParams) {
  return request.get<PaginatedResponse<Role>>('/api/identity/roles', { params });
}

/**
 * Get role by ID
 */
export function getRoleById(id: number) {
  return request.get<Role>(`/api/identity/roles/${id}`);
}

/**
 * Create role
 */
export function createRole(data: {
  name: string;
  code: string;
  description?: string;
}) {
  return request.post<{ id: number }>('/api/identity/roles', data);
}

/**
 * Update role
 */
export function updateRole(id: number, data: {
  name?: string;
  description?: string;
}) {
  return request.put<null>(`/api/identity/roles/${id}`, data);
}

/**
 * Delete role
 */
export function deleteRole(id: number) {
  return request.delete<null>(`/api/identity/roles/${id}`);
}

/**
 * Get role permissions
 */
export function getRolePermissions(roleId: number) {
  return request.get<{ permissionIds: number[] }>(`/api/identity/roles/${roleId}/permissions`);
}

/**
 * Assign permissions to role
 */
export function assignRolePermissions(roleId: number, permissionIds: number[]) {
  return request.post<null>(`/api/identity/roles/${roleId}/permissions`, {
    permission_ids: permissionIds
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
  return request.get<PaginatedResponse<HomeownerVerification>>('/api/identity/verifications', { params });
}

/**
 * Get verification by ID
 */
export function getVerificationById(id: number) {
  return request.get<HomeownerVerification>(`/api/identity/verifications/${id}`);
}

/**
 * Review homeowner verification
 */
export function reviewVerification(id: number, data: {
  status: number;
  review_notes?: string;
}) {
  return request.post<null>(`/api/identity/verifications/${id}/review`, data);
}
