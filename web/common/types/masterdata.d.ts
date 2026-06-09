// Masterdata Service types

export interface AdministrativeDivision {
  id: string;
  parent_id: string | null;
  level: DivisionLevel;
  name: string;
  code: string;
  path: string;
  sort_order: number;
  status: DivisionStatus;
  submission_status: number;
  submission_type: number | null;
  created_by: string;
  created_time: string;
  updated_time: string;
  children?: AdministrativeDivision[];
}

export enum DivisionLevel {
  Province = 1,
  City = 2,
  District = 3,
  Street = 4,
  Community = 5
}

export enum DivisionStatus {
  Active = 1,
  Inactive = 2
}

export interface ResidentialArea {
  id: string;
  county_id: string | null;
  street_id: string | null;
  community_div_id: string | null;
  code: string;
  name: string;
  address: string;
  area: number;
  population: number;
  community_type: CommunityType;
  submission_status: SubmissionStatus;
  submitter_id: string;
  submit_time: string;
  reviewer_id: string | null;
  review_time: string;
  review_notes: string;
  created_time: string;
  updated_time: string;
}

export enum CommunityType {
  Residential = 1,
  Village = 2,
  Mixed = 3
}

export enum SubmissionStatus {
  Draft = 0,
  Submitted = 1,
  Approved = 2,
  Rejected = 3,
  PendingDelete = 4
}

export interface Configuration {
  id: string;
  config_key: string;
  config_value: string;
  value_type: string; // string | number | boolean | json
  description: string;
  created_by: string;
  created_time: string;
  updated_time: string;
}

export enum ConfigValueType {
  String = 'string',
  Number = 'number',
  Boolean = 'boolean',
  Json = 'json'
}

export interface SensitiveWord {
  id: string;
  word: string;
  word_type: number;
  category: string;
  severity: Severity;
  action: SensitiveWordAction;
  status: SensitiveWordStatus;
  created_time: string;
  updated_time: string;
}

export enum Severity {
  Low = 1,
  Medium = 2,
  High = 3
}

export enum SensitiveWordAction {
  Warn = 1,
  Block = 2,
  Review = 3
}

export enum SensitiveWordStatus {
  Active = 1,
  Inactive = 2
}

export interface ResidentialAreaFilter {
  county_id?: string;
  street_id?: string;
  community_div_id?: string;
  submission_status?: SubmissionStatus;
  community_type?: CommunityType;
  keyword?: string;
}

export enum SubmissionType {
  Create = 1,
  Update = 2,
  Delete = 3
}

export enum EntityType {
  ResidentialArea = 'residential_area',
  AdministrativeDivision = 'administrative_division',
  SensitiveWord = 'sensitive_word'
}

export interface ApprovalPendingItem {
  id: string;
  entity_type: EntityType;
  submission_type: SubmissionType;
  name: string;
  change_summary: string;
  submitter_id: string;
  submit_time: string;
  submission_status: SubmissionStatus;
}

export interface PendingCounts {
  residential_area: number;
  administrative_division: number;
  sensitive_word: number;
  total: number;
}

export interface ApprovalDetail {
  id: string;
  entity_type: EntityType;
  submission_type: SubmissionType;
  current_data: string;
  snapshot_data: string;
  submitter_id: string;
  submit_time: string;
  reviewer_id: string | null;
  review_time: string;
  review_notes: string;
}

// 保留旧名称以兼容过渡期（后续清理）
/** @deprecated 使用 ResidentialArea 代替 */
export type Community = ResidentialArea
/** @deprecated 使用 ResidentialAreaFilter 代替 */
export type CommunityFilter = ResidentialAreaFilter

export interface ConfigFilter {
  keyword?: string;
}

export interface SensitiveWordFilter {
  category?: string;
  severity?: Severity;
  status?: SensitiveWordStatus;
}

export interface SubmissionRecord {
  id: string;
  entity_type: string;
  entity_id: string;
  entity_name: string;
  entity_code: string;
  submission_type: number;
  submitter_id: string;
  submit_time: string;
  reviewer_id: string | null;
  review_time: string;
  review_result: number;
  review_notes: string;
}

export interface DeletedItem {
  id: string;
  entity_type: string;
  name: string;
  code: string;
  delete_time: string;
}

export interface DeletedCounts {
  residential_area: number;
  administrative_division: number;
  sensitive_word: number;
  total: number;
}

export interface DivisionCountItem {
  id: string;
  name: string;
  level: number;
  community_count: number;
  village_count: number;
  total_count: number;
}
