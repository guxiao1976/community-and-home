import request from '@/utils/request'
import type { PaginatedResponse, PaginationParams } from '@common/types/common'
import type { AdministrativeDivision, ResidentialArea, ResidentialAreaFilter, ApprovalPendingItem, PendingCounts, ApprovalDetail, DeletedItem, DeletedCounts, DivisionCountItem } from '@common/types/masterdata'

// Get administrative divisions list
export const getAdministrativeDivisions = (params?: {
  parent_id?: number
  level?: number
  min_level?: number
  submission_status?: number
} & PaginationParams) => {
  return request.get<PaginatedResponse<AdministrativeDivision>>('/api/masterdata/divisions', { params })
}

// Get administrative division by ID
export const getAdministrativeDivisionById = (id: number) => {
  return request.get<AdministrativeDivision>(`/api/masterdata/divisions/${id}`)
}

// Create administrative division
export const createAdministrativeDivision = (data: {
  parent_id?: number | null
  level: number
  name: string
  code: string
  sort_order?: number
}) => {
  const payload: any = {
    level: data.level,
    name: data.name,
    code: data.code,
    sort_order: data.sort_order || 0
  }
  if (data.parent_id) {
    payload.parent_id = data.parent_id
  }
  return request.post<{ id: number }>('/api/masterdata/divisions', payload)
}

// Update administrative division
export const updateAdministrativeDivision = (id: number, data: {
  name?: string
  sort_order?: number
  status?: number
}) => {
  return request.put<void>(`/api/masterdata/divisions/${id}`, data)
}

// Delete administrative division
export const deleteAdministrativeDivision = (id: number) => {
  return request.delete<{ success: boolean }>(`/api/masterdata/divisions/${id}`)
}

// Check if division has child data
export const getDivisionChildCount = (id: number) => {
  return request.get<{ has_child_divisions: boolean; has_residential_areas: boolean; has_data: boolean }>(`/api/masterdata/divisions/${id}/child-count`)
}

// Submit division
export const submitDivision = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/divisions/${id}/submit`)
}

// Batch submit divisions
export const batchSubmitDivisions = (ids: number[]) => {
  return request.post<{ success: boolean }>('/api/masterdata/divisions/batch-submit', { ids })
}

// Request delete for approved division
export const requestDeleteDivision = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/divisions/${id}/request-delete`)
}

// Cancel delete request
export const cancelDeleteDivision = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/divisions/${id}/cancel-delete`)
}

// Withdraw submitted division
export const withdrawDivision = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/divisions/${id}/withdraw`)
}

// ==================== Residential Area Management ====================

// Get residential areas list with filters
export const getResidentialAreas = (params?: ResidentialAreaFilter & PaginationParams) => {
  return request.get<PaginatedResponse<ResidentialArea>>('/api/masterdata/residential-areas', { params })
}

// Get residential area by ID
export const getResidentialAreaById = (id: number) => {
  return request.get<ResidentialArea>(`/api/masterdata/residential-areas/${id}`)
}

// Create residential area
export const createResidentialArea = (data: {
  county_id: number
  street_id?: number
  community_div_id?: number
  code: string
  name: string
  address: string
  area?: number
  population?: number
  community_type: number
}) => {
  return request.post<{ id: number }>('/api/masterdata/residential-areas', data)
}

// Update residential area
export const updateResidentialArea = (id: number, data: {
  street_id?: number
  community_div_id?: number
  code?: string
  name?: string
  address?: string
  area?: number
  population?: number
  community_type?: number
}) => {
  return request.put<{ success: boolean }>(`/api/masterdata/residential-areas/${id}`, data)
}

// Submit residential area for review
export const submitResidentialArea = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/residential-areas/${id}/submit`)
}

// Batch submit residential areas
export const batchSubmitResidentialAreas = (ids: number[]) => {
  return request.post<{ success: boolean }>('/api/masterdata/residential-areas/batch-submit', { ids })
}

// Review residential area (approve/reject)
export const reviewResidentialArea = (id: number, data: {
  action: 'approve' | 'reject'
  review_notes?: string
}) => {
  return request.post<{ success: boolean }>(`/api/masterdata/residential-areas/${id}/review`, data)
}

// Delete residential area
export const deleteResidentialArea = (id: number) => {
  return request.delete<{ success: boolean }>(`/api/masterdata/residential-areas/${id}`)
}

// ==================== Configuration Management ====================

export const getConfigurations = (params?: {
  module?: string
  key?: string
} & PaginationParams) => {
  return request.get<PaginatedResponse<any>>('/api/masterdata/configurations', { params })
}

export const getConfigurationById = (id: number) => {
  return request.get<any>(`/api/masterdata/configurations/${id}`)
}

export const createConfiguration = (data: {
  module: string
  key: string
  value: string
  value_type: string
  description?: string
  is_public: boolean
}) => {
  return request.post<{ id: number }>('/api/masterdata/configurations', data)
}

export const updateConfiguration = (id: number, data: {
  value?: string
  description?: string
  is_public?: boolean
}) => {
  return request.put<null>(`/api/masterdata/configurations/${id}`, data)
}

export const deleteConfiguration = (id: number) => {
  return request.delete<null>(`/api/masterdata/configurations/${id}`)
}

export const submitConfiguration = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/configurations/${id}/submit`)
}

export const batchSubmitConfigurations = (ids: number[]) => {
  return request.post<{ success: boolean }>('/api/masterdata/configurations/batch-submit', { ids })
}

// ==================== Sensitive Word Management ====================

export const getSensitiveWords = (params?: {
  category?: string
  severity?: number
  status?: number
} & PaginationParams) => {
  return request.get<PaginatedResponse<any>>('/api/masterdata/sensitive-words', { params })
}

export const createSensitiveWord = (data: {
  word: string
  category: string
  severity: number
  action: string
}) => {
  return request.post<{ id: number }>('/api/masterdata/sensitive-words', data)
}

export const updateSensitiveWord = (id: number, data: {
  category?: string
  severity?: number
  action?: string
  status?: number
}) => {
  return request.put<null>(`/api/masterdata/sensitive-words/${id}`, data)
}

export const deleteSensitiveWord = (id: number) => {
  return request.delete<null>(`/api/masterdata/sensitive-words/${id}`)
}

export const submitSensitiveWord = (id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/sensitive-words/${id}/submit`)
}

export const batchSubmitSensitiveWords = (ids: number[]) => {
  return request.post<{ success: boolean }>('/api/masterdata/sensitive-words/batch-submit', { ids })
}

// ==================== Approval Center ====================

export const getPendingCounts = () => {
  return request.get<PendingCounts>('/api/masterdata/approval/pending-counts')
}

export const getPendingItems = (params?: {
  entity_type?: string
  submission_type?: number
  page?: number
  page_size?: number
}) => {
  return request.get<PaginatedResponse<ApprovalPendingItem>>('/api/masterdata/approval/pending-items', { params })
}

export const getApprovalDetail = (entityType: string, id: number) => {
  return request.get<ApprovalDetail>(`/api/masterdata/approval/${entityType}/${id}`)
}

export const reviewItem = (entityType: string, id: number, data: {
  action: 'approve' | 'reject'
  review_notes?: string
}) => {
  return request.post<{ success: boolean }>(`/api/masterdata/approval/${entityType}/${id}/review`, data)
}

export const batchReviewItems = (data: {
  entity_type: string
  ids: number[]
  action: 'approve' | 'reject'
  review_notes?: string
}) => {
  return request.post<{ success_count: number; fail_count: number }>('/api/masterdata/approval/batch-review', data)
}

// ==================== Submission Records ====================

export const getMySubmissionRecords = (params?: {
  entity_type?: string
  review_result?: number
  page?: number
  page_size?: number
}) => {
  return request.get<{ list: any[]; total: number }>('/api/masterdata/submission-records/my', { params })
}

export const getReviewedSubmissionRecords = (params?: {
  entity_type?: string
  review_result?: number
  page?: number
  page_size?: number
}) => {
  return request.get<{ list: any[]; total: number }>('/api/masterdata/submission-records/reviewed', { params })
}

// ==================== Deleted Data Recovery ====================

export const getDeletedCounts = () => {
  return request.get<DeletedCounts>('/api/masterdata/deleted-items/counts')
}

export const getDeletedItems = (params?: {
  entity_type?: string
  page?: number
  page_size?: number
}) => {
  return request.get<{ list: DeletedItem[]; total: number }>('/api/masterdata/deleted-items', { params })
}

export const restoreDeletedItem = (entityType: string, id: number) => {
  return request.post<{ success: boolean }>(`/api/masterdata/deleted-items/${entityType}/${id}/restore`)
}

// ==================== Statistics ====================

export const getDivisionCounts = (params?: { parent_id?: number }) => {
  return request.get<{ list: DivisionCountItem[] }>('/api/masterdata/statistics/division-counts', { params })
}

export const getDivisionCountsRealtime = (params?: { parent_id?: number }) => {
  return request.get<{ list: DivisionCountItem[] }>('/api/masterdata/statistics/division-counts/realtime', { params })
}

// ==================== AMap Sync ====================

export interface SyncProgress {
  task_id: string
  status: 'running' | 'completed' | 'failed'
  total_counties: number
  current_county: number
  current_county_name?: string
  total_streets: number
  current_street: number
  current_street_name?: string
  total_pages: number
  current_page: number
  total_found: number
  total_synced: number
  total_skipped: number
  total_failed: number
  error_message?: string
}

export const syncResidentialAreas = (data: { division_id: number }) => {
  return request.post<{ task_id: string }>('/api/masterdata/amap-sync/sync', data)
}

export const getSyncProgress = (taskId: string) => {
  return request.get<SyncProgress>('/api/masterdata/amap-sync/progress', {
    params: { task_id: taskId }
  })
}

// ==================== Data Query (Read-only) ====================

export interface QueryResidentialAreaItem {
  id: number
  code: string
  name: string
  address: string
  community_type: number
  city_id: number | null
  city_name: string
  county_id: number | null
  county_name: string
  street_id: number | null
  street_name: string
  community_div_id: number | null
  community_name: string
}

export const queryResidentialAreas = (params: {
  city_id?: number
  county_id?: number
  street_id?: number
  community_div_id?: number
  keyword?: string
  community_type?: number
  page?: number
  page_size?: number
}) => {
  return request.get<{ list: QueryResidentialAreaItem[]; total: number }>('/api/masterdata/query/residential-areas', { params })
}
