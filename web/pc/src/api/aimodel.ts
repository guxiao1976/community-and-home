// AI Model Service API

import request from '@/utils/request';
import type { PaginatedResponse, PaginationParams } from '@common/types/common';

/**
 * Model configuration
 */
export interface ModelConfig {
  id: number;
  name: string;
  display_name: string;
  provider: string;
  type: string;
  endpoint?: string;
  max_tokens: number;
  supported_features: string;
  cost_per_1k_input_tokens: number;
  cost_per_1k_output_tokens: number;
  status: number;
  description?: string;
  created_at: string;
  updated_at: string;
}

/**
 * API Key
 */
export interface ApiKey {
  id: number;
  model_id?: number;
  provider: string;
  key_name: string;
  api_key: string;
  status: number;
  created_at: string;
  updated_at: string;
}

/**
 * Prompt Template
 */
export interface PromptTemplate {
  id: number;
  name: string;
  content: string;
  category: string;
  created_at: string;
  updated_at: string;
}

/**
 * Model call request
 */
export interface ModelCallRequest {
  model_name: string;
  prompt: string;
  system_prompt?: string;
  max_tokens?: number;
  temperature?: number;
  parameters?: Record<string, any>;
}

/**
 * Model call response
 */
export interface ModelCallResponse {
  content: string;
  input_tokens: number;
  output_tokens: number;
  cost: number;
  latency: number;
  model_version: string;
  finish_reason: string;
}

/**
 * Health check record
 */
export interface HealthCheckRecord {
  id: number;
  model_id: number;
  status: number;
  response_time: number;
  error_message?: string;
  checked_time: string;
}

/**
 * Usage statistics
 */
export interface UsageStatistics {
  id: number;
  model_id: number;
  date: string;
  total_calls: number;
  success_calls: number;
  failed_calls: number;
  total_tokens: number;
  total_cost: number;
  avg_latency: number;
}

// ==================== Model Config APIs ====================

/**
 * Get model configs list
 */
export function getModelConfigs(params?: PaginationParams) {
  return request.get<PaginatedResponse<ModelConfig>>('/api/v1/models', { params });
}

/**
 * Get model config by ID
 */
export function getModelConfigById(id: number) {
  return request.get<ModelConfig>(`/api/v1/model/${id}`);
}

/**
 * Create model config
 */
export function createModelConfig(data: {
  name: string;
  display_name: string;
  provider: string;
  type: string;
  endpoint?: string;
  max_tokens: number;
  supported_features: string;
  cost_per_1k_input_tokens: number;
  cost_per_1k_output_tokens: number;
  description?: string;
}) {
  return request.post<{ id: number }>('/api/v1/model', data);
}

/**
 * Update model config
 */
export function updateModelConfig(data: {
  id: number;
  display_name?: string;
  endpoint?: string;
  max_tokens?: number;
  supported_features?: string;
  cost_per_1k_input_tokens?: number;
  cost_per_1k_output_tokens?: number;
  status?: number;
  description?: string;
}) {
  return request.put<null>('/api/v1/model', data);
}

/**
 * Delete model config
 */
export function deleteModelConfig(id: number) {
  return request.delete<null>(`/api/v1/model/${id}`);
}

// ==================== API Key APIs ====================

/**
 * Get API keys list
 */
export function getApiKeys(params?: PaginationParams) {
  return request.get<PaginatedResponse<ApiKey>>('/api/v1/apikeys', { params });
}

/**
 * Get API key by ID
 */
export function getApiKeyById(id: number) {
  return request.get<ApiKey>(`/api/v1/apikey/${id}`);
}

/**
 * Create API key
 */
export function createApiKey(data: {
  model_id: number;
  key_name: string;
  api_key: string;
  description?: string;
}) {
  return request.post<{ id: number }>('/api/v1/apikey', data);
}

/**
 * Update API key
 */
export function updateApiKey(data: {
  id: number;
  key_name?: string;
  api_key?: string;
}) {
  return request.put<null>('/api/v1/apikey', data);
}

/**
 * Delete API key
 */
export function deleteApiKey(id: number) {
  return request.delete<null>(`/api/v1/apikey/${id}`);
}

// ==================== Template APIs ====================

/**
 * Get templates list
 */
export function getTemplates(params?: PaginationParams) {
  return request.get<PaginatedResponse<PromptTemplate>>('/api/v1/templates', { params });
}

/**
 * Get template by ID
 */
export function getTemplateById(id: number) {
  return request.get<PromptTemplate>(`/api/v1/template/${id}`);
}

/**
 * Create template
 */
export function createTemplate(data: {
  name: string;
  content: string;
  category: string;
}) {
  return request.post<{ id: number }>('/api/v1/template', data);
}

/**
 * Update template
 */
export function updateTemplate(data: {
  id: number;
  name?: string;
  content?: string;
  category?: string;
}) {
  return request.put<null>('/api/v1/template', data);
}

/**
 * Delete template
 */
export function deleteTemplate(id: number) {
  return request.delete<null>(`/api/v1/template/${id}`);
}

// ==================== Model Call APIs ====================

/**
 * Call AI model
 */
export function callModel(data: ModelCallRequest) {
  return request.post<ModelCallResponse>('/api/v1/model/call', data);
}

// ==================== Health Check APIs ====================

/**
 * Get health check records
 */
export function getHealthChecks(params?: {
  model_id?: number;
  start_date?: string;
  end_date?: string;
} & PaginationParams) {
  return request.get<PaginatedResponse<HealthCheckRecord>>('/api/v1/health-checks', { params });
}

/**
 * Trigger health check for a model
 */
export function triggerHealthCheck(modelConfigId: number) {
  return request.post<HealthCheckRecord>(`/api/v1/model/${modelConfigId}/health-check`);
}

// ==================== Statistics APIs ====================

/**
 * Get usage statistics
 */
export function getUsageStatistics(params?: {
  model_id?: number;
  start_date?: string;
  end_date?: string;
} & PaginationParams) {
  return request.get<PaginatedResponse<UsageStatistics>>('/api/v1/statistics', { params });
}
