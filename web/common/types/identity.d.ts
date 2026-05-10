// Identity Service types

export interface User {
  id: number;
  phone: string;
  nickname: string;
  avatar: string;
  userType: UserType;
  status: UserStatus;
  verificationStatus: VerificationStatus;
  scope: string;
  lastLoginAt: string;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

export enum UserType {
  Staff = 1,
  Homeowner = 2
}

export enum UserStatus {
  Active = 1,
  Disabled = 2,
  Locked = 3
}

export enum VerificationStatus {
  Unverified = 0,
  Verified = 1,
  Rejected = 2
}

export interface Role {
  id: number;
  name: string;
  code: string;
  description: string;
  isSystem: boolean;
  status: RoleStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

export enum RoleStatus {
  Active = 1,
  Disabled = 2
}

export interface Permission {
  id: number;
  parentId: number;
  name: string;
  code: string;
  type: PermissionType;
  path: string;
  icon: string;
  sortOrder: number;
  status: PermissionStatus;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  children?: Permission[];
}

export enum PermissionType {
  Menu = 1,
  Button = 2
}

export enum PermissionStatus {
  Active = 1,
  Disabled = 2
}

export interface HomeownerVerification {
  id: number;
  userId: number;
  propertyUnit: string;
  realName: string;
  idCard: string;
  documentUrls: string[];
  verificationStatus: HomeownerVerificationStatus;
  reviewerId: number;
  reviewedAt: string;
  reviewNotes: string;
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
  user?: User;
  reviewer?: User;
}

export enum HomeownerVerificationStatus {
  Pending = 0,
  Approved = 1,
  Rejected = 2
}

export interface LoginRequest {
  phone: string;
  password: string;
}

export interface LoginSmsRequest {
  phone: string;
  smsCode: string;
}

export interface RegisterRequest {
  phone: string;
  password?: string;
  smsCode: string;
  nickname: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
}

export interface UserFilter {
  userType?: UserType;
  status?: UserStatus;
  verificationStatus?: VerificationStatus;
  keyword?: string;
}
