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
  check_mode?: 'ac_only' | 'model_only' | 'combined';
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
