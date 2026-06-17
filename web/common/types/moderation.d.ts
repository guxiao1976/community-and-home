// Traditional technology check result
export interface TraditionalCheckResult {
  passed: boolean;
  reason?: string;
  keywords?: string[];
  score?: number;
}

// Small model check result
export interface SmallModelCheckResult {
  passed: boolean;
  confidence: number;
  categories?: string[];
  reason?: string;
}

// Large model check result
export interface LargeModelCheckResult {
  passed: boolean;
  confidence: number;
  analysis?: string;
  categories?: string[];
  reason?: string;
}

// Text moderation request
export interface TextModerationRequest {
  content: string;
  userId?: string;
  scene?: string;
  check_mode?: 'ac_only' | 'small_model' | 'large_model' | 'combined';
}

// Text moderation response
export interface TextModerationResponse {
  requestId: string;
  finalResult: boolean;
  traditional: TraditionalCheckResult;
  smallModel: SmallModelCheckResult;
  largeModel?: LargeModelCheckResult;
  processingTime: number;
}

// Image moderation request
export interface ImageModerationRequest {
  imageBase64: string;
  userId?: string;
  scene?: string;
}

// Image moderation response
export interface ImageModerationResponse {
  requestId: string;
  finalResult: boolean;
  smallModel: SmallModelCheckResult;
  largeModel?: LargeModelCheckResult;
  processingTime: number;
}

// ========== 管线配置类型 ==========

export interface PipelineConfig {
  id: string;
  pipeline_id: string;
  pipeline_name: string;
  description: string;
  is_active: number;
  is_production: number;
  ac_enabled: number;
  ac_severity_threshold: number;
  small_model_template_id: string;
  small_model_config_key: string;
  large_model_template_id: string;
  large_model_config_key: string;
  ac_to_small_condition: string;
  ac_to_small_severity: number;
  ac_to_small_categories: string[];
  small_to_large_condition: string;
  small_to_large_confidence_threshold: number;
  small_to_large_categories: string[];
  final_verdict_logic: string;
  created_time: string;
  updated_time: string;
}

export interface PipelineListResponse {
  list: PipelineConfig[];
  total: number;
  page: number;
  page_size: number;
}

export interface CreatePipelineRequest {
  pipeline_id: string;
  pipeline_name: string;
  description?: string;
  ac_enabled?: number;
  ac_severity_threshold?: number;
  small_model_template_id?: string;
  small_model_config_key?: string;
  large_model_template_id?: string;
  large_model_config_key?: string;
  ac_to_small_condition?: string;
  ac_to_small_severity?: number;
  ac_to_small_categories?: string[];
  small_to_large_condition?: string;
  small_to_large_confidence_threshold?: number;
  small_to_large_categories?: string[];
  final_verdict_logic?: string;
}

export interface UpdatePipelineRequest {
  id: string;
  pipeline_name?: string;
  description?: string;
  is_active?: number;
  is_production?: number;
  ac_enabled?: number;
  ac_severity_threshold?: number;
  small_model_template_id?: string;
  small_model_config_key?: string;
  large_model_template_id?: string;
  large_model_config_key?: string;
  ac_to_small_condition?: string;
  ac_to_small_severity?: number;
  ac_to_small_categories?: string[];
  small_to_large_condition?: string;
  small_to_large_confidence_threshold?: number;
  small_to_large_categories?: string[];
  final_verdict_logic?: string;
}

// ========== 管线测试类型 ==========

export interface PipelineTestRequest {
  content: string;
  pipeline_id?: string;
  ac_enabled?: number;
  ac_severity_threshold?: number;
  small_model_template_id?: string;
  large_model_template_id?: string;
  ac_to_small_condition?: string;
  ac_to_small_severity?: number;
  ac_to_small_categories?: string[];
  small_to_large_condition?: string;
  small_to_large_confidence_threshold?: number;
  small_to_large_categories?: string[];
  final_verdict_logic?: string;
}

export interface PipelineLayerResult {
  called: boolean;
  skipped_reason?: string;
  passed: boolean;
  risk_level: string;
  confidence: number;
  categories?: string[];
  reason?: string;
  latency_ms: number;
  matched_words?: string[];
  model_used?: string;
  template_id?: string;
  raw_response?: string;
}

export interface PipelineTestResponse {
  pipeline_id: string;
  ac_result: PipelineLayerResult | null;
  small_model_result: PipelineLayerResult | null;
  large_model_result: PipelineLayerResult | null;
  final_verdict: string;
  total_latency_ms: number;
}

// ========== 人工审核类型 ==========

export interface ReviewListItem {
  id: string;
  source_type: string;
  source_id: string;
  content_summary: string;
  risk_level: string;
  pass: boolean;
  review_status: number;
  created_time: string;
}

export interface ReviewDetail {
  id: string;
  source_type: string;
  source_id: string;
  content_type: string;
  content_summary: string;
  risk_level: string;
  pass: boolean;
  reason: string;
  check_layer: string;
  matched_items: string;
  review_status: number;
  review_notes: string;
  created_time: string;
}

export interface ReviewListResponse {
  list: ReviewListItem[];
  total: number;
  page: number;
  page_size: number;
}

export interface ReviewListParams {
  source_type?: string;
  review_status: number;
  page?: number;
  page_size?: number;
}

export interface SubmitReviewRequest {
  audit_log_id: string;
  review_status: number;
  review_notes?: string;
}

// source_type display name mapping
export const SOURCE_TYPE_LABELS: Record<string, string> = {
  notice: '通知公告',
  lost_found: '寻失互助',
  certification: '房主认证',
  nickname: '用户昵称',
};

// moderation_status display mapping
export const MODERATION_STATUS_MAP: Record<number, { label: string; type: string }> = {
  0: { label: '待审核', type: 'info' },
  1: { label: '机器通过', type: 'success' },
  2: { label: '机器不通过', type: 'danger' },
  3: { label: '人审通过', type: '' },
  4: { label: '人审不通过', type: 'danger' },
};
