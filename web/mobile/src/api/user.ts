// User Service & Master Data API — Uni-app mobile
import request from '@/utils/request';

// CommunityMembership 复用 web/common（字段名已与 wire snake_case 对齐），不再端内重定义。
// 本地别名指向共享层（避免 vue-tsc 对「import type + 使用」解析问题），仍是单一共享类型源。
// SEE: [[web-common-type-reuse-no-redefine]] [[web-common-type-field-wire-mismatch]]
import type { CommunityMembership as CMembership } from '@common/types/identity';
type CommunityMembership = CMembership;

export interface Division {
  id: string;
  name: string;
  level: number; // 1=province, 2=city, 3=district, 4=street, 5=community
  parentId?: string | null;
}

export interface ResidentialArea {
  id: string;
  name: string;
  address: string;
  countyId?: string;
  countyName?: string;
  cityName?: string;
}

/**
 * Join a community. Requires JWT.
 * 新模型：加入=立即建 membership（无房号），房号/权属绑定移到独立步骤（bindResidence + applyRole），
 * 故 building/unit/room/ownership 均改为可选——仅调用方传入时才透传给后端（并行管线已支持房号可选）。
 * 传入时与 proto JoinCommunityRequest 对齐（ownership: 1=OWNED(自有)→owner / 2=RENTED(租住)→tenant）。
 */
export async function joinCommunity(
  communityId: string,
  building?: number,
  unit?: number,
  room?: number,
  ownership?: number,
): Promise<CommunityMembership> {
  const payload: Record<string, unknown> = { community_id: communityId };
  if (building != null) payload.building = building;
  if (unit != null) payload.unit = unit;
  if (room != null) payload.room = room;
  if (ownership != null) payload.ownership = ownership;
  const res = await request.post<any>('/api/users/communities/join', payload);
  // REST 响应经拦截器解包 data 后仍是 { membership: {...} } 包装对象；须解出 membership 资源本身，
  // 否则调用方取 res.id 恒为 undefined（membershipId 回填静默失效，靠 getUserMemberships 兜底掩盖）。
  // SEE: [[frontend-api-return-wrapped-resource-unwrap]] [[api-response-single-wrap]]
  const data = res as { membership?: CommunityMembership };
  return data.membership || (res as unknown as CommunityMembership);
}

/**
 * Leave a community. Requires JWT.
 */
export async function leaveCommunity(communityId: string): Promise<void> {
  await request.post('/api/users/communities/leave', {
    community_id: communityId,
  });
}

/**
 * Get user's active community memberships. Requires JWT.
 */
export async function getUserMemberships(): Promise<CommunityMembership[]> {
  const res = await request.get<{ memberships: CommunityMembership[] }>(
    '/api/users/communities/memberships',
  );
  const data = res as unknown as { memberships: CommunityMembership[] };
  return data.memberships || [];
}

/**
 * Get divisions. If parentId is not provided, returns top-level (provinces).
 * Backend: GET /api/masterdata/divisions[?parent_id=xxx]
 * Response: { list: [{ id, name, level, parentId?, ... }], total }
 */
export async function getDivisions(parentId?: string): Promise<Division[]> {
  const params = parentId ? `?parent_id=${parentId}` : '';
  const res = await request.get<any>(`/api/masterdata/divisions${params}`);
  const data = res as unknown as { list?: Division[] };
  return data.list || [];
}

/**
 * Search residential areas (communities) by county/district.
 * Backend: GET /api/masterdata/query/residential-areas?county_id=xxx&keyword=xxx
 * Response: { list: [{ id, name, address, countyName, cityName, ... }], total }
 */
export async function searchResidentialAreas(params: {
  keyword?: string;
  countyId?: string;
}): Promise<ResidentialArea[]> {
  const query = new URLSearchParams();
  if (params.keyword) query.set('keyword', params.keyword);
  if (params.countyId) query.set('county_id', params.countyId);
  const res = await request.get<any>(
    `/api/masterdata/query/residential-areas?${query.toString()}`,
  );
  const data = res as unknown as { list?: ResidentialArea[] };
  return data.list || [];
}

/**
 * Bind a residence (building/unit/room) to a membership.
 * Required before applying for owner/tenant role.
 */
export async function bindResidence(params: {
  membership_id: string;
  building: string;
  unit: string;
  room: string;
  is_primary?: number;
  start_date?: string;
  end_date?: string;
}): Promise<any> {
  const res = await request.post('/api/users/residences/bind', params);
  return res;
}

/**
 * Apply for a role (owner, tenant, grid_worker, etc.)
 * 后端走 permission-service 写入 rel_user_role（角色申请，status=0 未认证）
 */
export async function applyRole(params: {
  community_id: string;
  role_code: string;
}): Promise<any> {
  const res = await request.post('/api/users/roles/apply', params);
  return res;
}

/**
 * Get residential areas by IDs (batch query).
 */
export async function getResidentialAreasByIds(ids: string[]): Promise<ResidentialArea[]> {
  const res = await request.post<any>('/api/masterdata/residential-areas/batch', { ids });
  const data = res as unknown as { list?: ResidentialArea[] };
  return data.list || [];
}

/**
 * Get user's roles.
 * 后端从 permission-service 获取（含认证状态 status）
 */
export async function getUserRoles(): Promise<any[]> {
  const res = await request.get<any>('/api/users/roles');
  const data = res as unknown as { roles?: any[] };
  return data.roles || [];
}

// 同屋互见信息 — matches proto SameHouseInfo（snake_case over the wire）。
// same_house=true 时后端返回真实手机号 + 楼/单元/房号；否则手机号脱敏且不返回房屋号。
export interface SameHouseInfo {
  same_house: boolean;
  building: number;
  unit: number;
  room: number;
}

// GetUser 响应（同屋互见）— matches proto GetUserResponse。
// phone 已由后端按 viewer 上下文决定明文/脱敏，前端仅展示、不做二次脱敏（权威在后端）。
export interface GetUserDetailResult {
  user: {
    id: string;
    phone: string;
    nickname: string;
  };
  same_house?: SameHouseInfo;
}

/**
 * Get a user detail with viewer context (同屋互见).
 * viewer_id 缺省（0/不传）→ 后端对手机号脱敏、无房屋号。
 * 后端: GET /api/users/:id?viewer_id=xxx
 */
export async function getUser(id: string, viewerId?: string): Promise<GetUserDetailResult> {
  const query = viewerId ? `?viewer_id=${viewerId}` : '';
  const res = await request.get<any>(`/api/users/${id}${query}`);
  return res as unknown as GetUserDetailResult;
}

/**
 * Get current community app state (跨设备一致).
 * 后端: GET /api/users/me/app-state
 * Response: { current_community_id, updated_at }（未设置 current_community_id="0"）
 */
export async function getAppState(): Promise<{
  current_community_id: string;
  updated_at: number;
}> {
  const res = await request.get<any>('/api/users/me/app-state');
  return res as unknown as { current_community_id: string; updated_at: number };
}

/**
 * Switch current community (跨设备持久化).
 * 后端: PUT /api/users/me/current-community，失败 10015（目标小区不在数据范围）。
 */
export async function setCurrentCommunity(communityId: string): Promise<void> {
  await request.put('/api/users/me/current-community', { community_id: communityId });
}
