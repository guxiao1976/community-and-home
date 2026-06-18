// Common API types shared across all services

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export interface ApiError {
  code: number;
  message: string;
}

export interface PageRequest {
  page: number;
  pageSize: number;
}

export interface PaginationParams {
  page?: number;
  page_size?: number;
}

export interface PageResponse<T> {
  list: T[];
  total: number;
}

export interface PaginatedResponse<T> extends PageResponse<T> {
  page: number;
  pageSize: number;
}
