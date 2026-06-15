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
  small_model_template_id?: string;
  large_model_template_id?: string;
  small_to_large_confidence?: number;
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
