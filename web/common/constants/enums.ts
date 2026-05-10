// Business enums with display labels

import {
  UserType,
  UserStatus,
  VerificationStatus,
  RoleStatus,
  PermissionType,
  PermissionStatus,
  HomeownerVerificationStatus
} from '../types/identity';

import {
  DivisionLevel,
  DivisionStatus,
  CommunityType,
  SubmissionStatus,
  ConfigValueType,
  ApprovalStatus,
  Severity,
  SensitiveWordAction,
  SensitiveWordStatus
} from '../types/masterdata';

// Re-export enums
export {
  UserType,
  UserStatus,
  VerificationStatus,
  RoleStatus,
  PermissionType,
  PermissionStatus,
  HomeownerVerificationStatus,
  DivisionLevel,
  DivisionStatus,
  CommunityType,
  SubmissionStatus,
  ConfigValueType,
  ApprovalStatus,
  Severity,
  SensitiveWordAction,
  SensitiveWordStatus
};

// Display labels and colors for UI

export const USER_TYPE_LABELS: Record<UserType, string> = {
  [UserType.Staff]: '员工',
  [UserType.Homeowner]: '业主'
};

export const USER_STATUS_LABELS: Record<UserStatus, { label: string; color: string }> = {
  [UserStatus.Active]: { label: '正常', color: 'success' },
  [UserStatus.Disabled]: { label: '禁用', color: 'danger' },
  [UserStatus.Locked]: { label: '锁定', color: 'warning' }
};

export const VERIFICATION_STATUS_LABELS: Record<VerificationStatus, { label: string; color: string }> = {
  [VerificationStatus.Unverified]: { label: '未认证', color: 'info' },
  [VerificationStatus.Verified]: { label: '已认证', color: 'success' },
  [VerificationStatus.Rejected]: { label: '已拒绝', color: 'danger' }
};

export const HOMEOWNER_VERIFICATION_STATUS_LABELS: Record<HomeownerVerificationStatus, { label: string; color: string }> = {
  [HomeownerVerificationStatus.Pending]: { label: '待审核', color: 'warning' },
  [HomeownerVerificationStatus.Approved]: { label: '已通过', color: 'success' },
  [HomeownerVerificationStatus.Rejected]: { label: '已拒绝', color: 'danger' }
};

export const DIVISION_LEVEL_LABELS: Record<DivisionLevel, string> = {
  [DivisionLevel.Province]: '省',
  [DivisionLevel.City]: '市',
  [DivisionLevel.District]: '区',
  [DivisionLevel.Street]: '街道',
  [DivisionLevel.Community]: '社区'
};

export const COMMUNITY_TYPE_LABELS: Record<CommunityType, string> = {
  [CommunityType.Residential]: '住宅小区',
  [CommunityType.Village]: '农村社区',
  [CommunityType.Mixed]: '混合社区'
};

export const SUBMISSION_STATUS_LABELS: Record<SubmissionStatus, { label: string; color: string }> = {
  [SubmissionStatus.Draft]: { label: '草稿', color: 'info' },
  [SubmissionStatus.Submitted]: { label: '待审核', color: 'warning' },
  [SubmissionStatus.Approved]: { label: '已通过', color: 'success' },
  [SubmissionStatus.Rejected]: { label: '已拒绝', color: 'danger' }
};

export const SEVERITY_LABELS: Record<Severity, { label: string; color: string }> = {
  [Severity.Low]: { label: '低', color: 'info' },
  [Severity.Medium]: { label: '中', color: 'warning' },
  [Severity.High]: { label: '高', color: 'danger' }
};

export const SENSITIVE_WORD_ACTION_LABELS: Record<SensitiveWordAction, string> = {
  [SensitiveWordAction.Warn]: '警告',
  [SensitiveWordAction.Block]: '拦截',
  [SensitiveWordAction.Review]: '人工审核'
};
