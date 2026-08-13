// AI Model Service API

import request from '@/utils/request';
import type { PaginatedResponse, PaginationParams } from '@common/types/common';

/**
 * Model configuration
 */
export interface ModelConfig {
  id: string;
  name: string;
  config_key?: string;
  display_name: string;
  provider: string;
  type: string;
  model_type?: string;
  endpoint?: string;
  max_tokens: number;
  temperature: number;
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
  id: string;
  model_id: string;
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
  id: string;
  name: string;
  config_key?: string;
  model_name: string;
  content: string;
  category: string;
  variables: string[];
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
  id: string;
  model_id: string;
  status: number;
  response_time: number;
  error_message?: string;
  checked_time: string;
}

/**
 * Usage statistics
 */
export interface UsageStatistics {
  id: string;
  model_id: string;
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
  return request.get<{ models: ModelConfig[]; total: number }>('/api/v1/models', { params });
}

/**
 * Get model config by ID
 */
export function getModelConfigById(id: string) {
  return request.get<ModelConfig>(`/api/v1/model/${id}`);
}

/**
 * Create model config
 */
export function createModelConfig(data: {
  name: string;
  model_name?: string;
  config_key?: string;
  display_name: string;
  provider: string;
  type: string;
  model_type?: string;
  endpoint?: string;
  max_tokens: number;
  temperature?: number;
  supported_features: string;
  cost_per_1k_input_tokens: number;
  cost_per_1k_output_tokens: number;
  description?: string;
}) {
  return request.post<{ id: string }>('/api/v1/model', data);
}

/**
 * Update model config
 */
export function updateModelConfig(data: {
  id: string;
  name?: string;
  config_key?: string;
  display_name?: string;
  provider?: string;
  type?: string;
  model_type?: string;
  endpoint?: string;
  max_tokens?: number;
  temperature?: number;
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
export function deleteModelConfig(id: string) {
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
export function getApiKeyById(id: string) {
  return request.get<ApiKey>(`/api/v1/apikey/${id}`);
}

/**
 * Create API key
 */
export function createApiKey(data: {
  model_id: string;
  key_name: string;
  api_key: string;
  description?: string;
}) {
  return request.post<{ id: string }>('/api/v1/apikey', data);
}

/**
 * Update API key
 */
export function updateApiKey(data: {
  id: string;
  key_name?: string;
  api_key?: string;
}) {
  return request.put<null>('/api/v1/apikey', data);
}

/**
 * Delete API key
 */
export function deleteApiKey(id: string) {
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
export function getTemplateById(id: string) {
  return request.get<PromptTemplate>(`/api/v1/template/${id}`);
}

/**
 * Create template
 */
export function createTemplate(data: {
  name: string;
  config_key: string;
  model_name?: string;
  content: string;
  category: string;
}) {
  return request.post<{ id: string }>('/api/v1/template', data);
}

/**
 * Update template
 */
export function updateTemplate(data: {
  id: string;
  name?: string;
  config_key?: string;
  model_name?: string;
  content?: string;
  category?: string;
}) {
  return request.put<null>('/api/v1/template', data);
}

/**
 * Delete template
 */
export function deleteTemplate(id: string) {
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
  model_id: string;
  start_date?: string;
  end_date?: string;
} & PaginationParams) {
  return request.get<PaginatedResponse<HealthCheckRecord>>('/api/v1/health-checks', { params });
}

/**
 * Trigger health check for a model
 */
export function triggerHealthCheck(modelConfigId: string) {
  return request.post<HealthCheckRecord>(`/api/v1/model/${modelConfigId}/health-check`);
}

// ==================== Statistics APIs ====================

/**
 * Get usage statistics
 */
export function getUsageStatistics(params?: {
  model_id: string;
  start_date?: string;
  end_date?: string;
} & PaginationParams) {
  return request.get<{ statistics: UsageStatistics[]; total: number }>('/api/v1/statistics', { params });
}

// ==================== Test Connection API ====================

/** Test model connection */
export function testModelConnection(data: {
  model_name: string;
  endpoint: string;
  api_key: string;
  provider?: string;
}) {
  return request.post<{ success: boolean; latency_ms: number; error?: string }>(
    '/api/v1/model/test-connection', data
  );
}

// ==================== Test Template API ====================

/** Test prompt template with variable values */
export function testTemplate(data: {
  model_name: string;
  content: string;
  variables: Record<string, string>;
}) {
  return request.post<{
    rendered: string;
    unreplaced_vars: string[];
    response: string;
    input_tokens: number;
    output_tokens: number;
    cost: number;
    latency_ms: number;
  }>('/api/v1/template/test', data);
}

/**
 * 从供应商端点获取可用模型列表
 */
export function fetchProviderModels(data: {
  provider: string;
  endpoint: string;
  api_key?: string;
}) {
  return request.post<Array<{ model_name: string; display_name: string }>>(
    '/api/v1/model/fetch-provider-models',
    data
  );
}

// ========== 审核管线辅助 API ==========

/** 获取审核类型的模板列表（用于下拉选择） */
export function getModerationTemplates() {
  return request.get<PaginatedResponse<PromptTemplate>>('/api/v1/templates', {
    params: { category: 'moderation', page: 1, page_size: 50 }
  });
}

/** 获取健康可用的模型列表（用于下拉选择） */
export function getAvailableModels() {
  return request.get<{ models: ModelConfig[] }>('/api/v1/models');
}
