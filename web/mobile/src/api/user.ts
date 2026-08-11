// User Service & Master Data API — Uni-app mobile
import request from '@/utils/request';

// CommunityMembership — matches proto CommunityMembership (snake_case over the wire).
// protojson outputs snake_case by default; CommunityMembership only contains IDs
// (no community_name); use CommunityStore for display names.
export interface CommunityMembership {
  id: string;
  user_id: string;
  community_id: string;
  bind_status: number;
  join_time: number;
  leave_time: number;
  created_at: number;
  updated_at: number;
  building: number;
  unit: number;
  room: number;
}

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
 */
export async function joinCommunity(communityId: string): Promise<CommunityMembership> {
  const res = await request.post<CommunityMembership>(
    '/api/users/communities/join',
    { community_id: communityId },
  );
  return res as unknown as CommunityMembership;
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
