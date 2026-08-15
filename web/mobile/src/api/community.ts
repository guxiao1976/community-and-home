// Community Hub Service API — Uni-app mobile
import request from '@/utils/request';

// =============================================================================
// Type Definitions
// =============================================================================

// 字段名对齐后端 community-hub-service types.go JSON tag（snake_case）
// SEE: [[snake-camel-field-mismatch]]
export interface NoticeAttachment {
  id: string;
  file_name: string;
  file_url: string;
  file_size: number;
}

export interface Notice {
  id: string;
  community_id: string;
  title: string;
  content: string;
  role: number;
  publisher: string;
  publisher_id: string;
  is_pinned: boolean;
  published_at: number;
  created_at: number;
  updated_at: number;
  attachments: NoticeAttachment[];
}

export interface Contact {
  id: string;
  community_id: string;
  category: number;
  name: string;
  phone: string;
  sort_order: number;
}

export interface LostFoundItem {
  id: string;
  community_id: string;
  type: number;
  title: string;
  description: string;
  image_urls: string[];
  contact_phone: string;
  status: number;
  publisher_id: string;
  created_at: number;
}

// =============================================================================
// Enum Display Helpers
// =============================================================================

export function getNoticeRoleName(role: number): string {
  switch (role) {
    case 1: return '社区';
    case 2: return '业委会';
    case 3: return '物业';
    case 4: return '网格员';
    default: return '';
  }
}

export function getNoticeRoleColor(role: number): string {
  switch (role) {
    case 1: return '#B8956A';
    case 2: return '#8DAF7E';
    case 3: return '#D4958A';
    case 4: return '#E8C98E';
    default: return '#A6988A';
  }
}

export function getContactCategoryName(category: number): string {
  switch (category) {
    case 1: return '供水维修';
    case 2: return '电力维修';
    case 3: return '燃气维修';
    case 4: return '联通网络';
    case 5: return '移动网络';
    case 6: return '电信网络';
    case 7: return '小区民警';
    default: return '';
  }
}

export function getContactCategoryIcon(category: number): string {
  switch (category) {
    case 1: return '💧';
    case 2: return '⚡';
    case 3: return '🔥';
    case 4: return '📶';
    case 5: return '📱';
    case 6: return '🌐';
    case 7: return '👮';
    default: return '📞';
  }
}

export function getLostFoundTypeName(type: number): string {
  switch (type) {
    case 1: return '寻物';
    case 2: return '招领';
    default: return '';
  }
}

// =============================================================================
// API Functions
// =============================================================================

/**
 * Get notice list for a community.
 */
export async function getNoticeList(
  communityId: string,
  page: number = 1,
  pageSize: number = 3,
): Promise<{ notices: Notice[]; total: string }> {
  const res = await request.get('/api/community/notices', {
    params: { community_id: communityId, page, page_size: pageSize },
  });
  return res as unknown as { notices: Notice[]; total: string };
}

/**
 * Get a single notice detail by ID.
 */
export async function getNoticeDetail(id: string): Promise<{ notice: Notice }> {
  const res = await request.get(`/api/community/notices/${id}`);
  return res as unknown as { notice: Notice };
}

/**
 * Get contact list for a community.
 */
export async function getContacts(
  communityId: string,
): Promise<{ contacts: Contact[] }> {
  const res = await request.get('/api/community/contacts', {
    params: { community_id: communityId },
  });
  return res as unknown as { contacts: Contact[] };
}

/**
 * Get lost & found list for a community.
 * SEE: [[verify-api-before-calling]] — 路径对齐后端已注册路由（无连字符 lostfound）
 */
export async function getLostFoundList(
  communityId: string,
  page: number = 1,
  pageSize: number = 3,
): Promise<{ items: LostFoundItem[]; total: string }> {
  const res = await request.get('/api/community/lostfound', {
    params: { community_id: communityId, page, page_size: pageSize },
  });
  return res as unknown as { items: LostFoundItem[]; total: string };
}
