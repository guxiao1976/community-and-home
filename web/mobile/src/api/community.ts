// Community Hub Service API — Uni-app mobile
import request from '@/utils/request';

// =============================================================================
// Type Definitions
// =============================================================================

export interface NoticeAttachment {
  id: string;
  fileName: string;
  fileUrl: string;
  fileSize: number;
}

export interface Notice {
  id: string;
  communityId: string;
  title: string;
  content: string;
  role: number;
  publisher: string;
  publisherId: string;
  isPinned: boolean;
  publishedAt: number;
  createdAt: number;
  updatedAt: number;
  attachments: NoticeAttachment[];
}

export interface Contact {
  id: string;
  communityId: string;
  category: number;
  name: string;
  phone: string;
  sortOrder: number;
}

export interface LostFoundItem {
  id: string;
  communityId: string;
  type: number;
  title: string;
  description: string;
  imageUrls: string[];
  contactPhone: string;
  status: number;
  publisherId: string;
  createdAt: number;
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
 */
export async function getLostFoundList(
  communityId: string,
  page: number = 1,
  pageSize: number = 3,
): Promise<{ items: LostFoundItem[]; total: string }> {
  const res = await request.get('/api/community/lost-found', {
    params: { community_id: communityId, page, page_size: pageSize },
  });
  return res as unknown as { items: LostFoundItem[]; total: string };
}
