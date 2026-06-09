// Identity Service types

export interface User {
  id: string;
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
  id: string;
  name: string;
  code: string;
  description: string;
  isSystem: boolean;
  status: RoleStatus;
  sortOrder: number;
  permissions: Permission[];
  createdAt: string;
  updatedAt: string;
  deleteTime: number;
}

export enum RoleStatus {
  Active = 1,
  Disabled = 2
}

export interface Permission {
  id: string;
  parentId: string;
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
  id: string;
  userId: string;
  propertyUnit: string;
  realName: string;
  idCard: string;
  documentUrls: string[];
  verificationStatus: HomeownerVerificationStatus;
  reviewerId: string;
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
  encryptedPhone: string;
  encryptedPassword: string;
  deviceId: string;
  deviceType?: string;
}

export interface LoginSmsRequest {
  encryptedPhone: string;
  smsCode: string;
  deviceId: string;
  deviceType?: string;
}

export interface RegisterRequest {
  encryptedPhone: string;
  encryptedPassword?: string;
  smsCode: string;
  nickname: string;
  deviceId: string;
  deviceType?: string;
}

export interface LoginResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  userId: string;
}

export interface RefreshTokenRequest {
  refreshToken: string;
}

export interface RefreshTokenResponse {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

export interface UserFilter {
  userType?: UserType;
  status?: UserStatus;
  verificationStatus?: VerificationStatus;
  keyword?: string;
}

// Community membership — user's affiliation with a residential community
export interface CommunityMembership {
  id: string;
  userId: string;
  communityId: string;
  bindStatus: number;
  joinTime: number;
  leaveTime: number;
  createdAt: number;
  updatedAt: number;
  building: number;
  unit: number;
  room: number;
}

// Certification — user identity verification record
export interface Certification {
  id: string;
  roleId: string;
  userId: string;
  documentUrls: string;
  status: number;
  reviewerId: string;
  reviewNotes: string;
  reviewTime: number;
  submitTime: number;
}

// Residence — user's registered dwelling within a community
export interface Residence {
  id: string;
  membershipId: string;
  houseId: string;
  building: string;
  unit: string;
  room: string;
  isPrimary: number;
  startDate: string;
  endDate: string;
  createdAt: number;
  updatedAt: number;
}
