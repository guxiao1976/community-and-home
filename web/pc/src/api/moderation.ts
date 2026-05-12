import request from '@/utils/request';
import type {
  TextModerationRequest,
  TextModerationResponse,
  ImageModerationRequest,
  ImageModerationResponse
} from '@common/types/moderation';

// Backend response structure
interface BackendMatchDetail {
  layer: string;
  matched_text?: string;
  category?: string;
  severity?: number;
  confidence?: number;
}

interface BackendModerationResponse {
  pass: boolean;
  risk_level: string;
  reason: string;
  need_review: boolean;
  details: BackendMatchDetail[];
}

/**
 * Convert backend flat response to frontend layered structure
 */
function convertTextResponse(backend: BackendModerationResponse): TextModerationResponse {
  const startTime = Date.now();

  // Group details by layer
  const acDetails = backend.details.filter(d => d.layer === 'ac_engine');
  const smallModelDetails = backend.details.filter(d => d.layer === 'small_model');
  const largeModelDetails = backend.details.filter(d => d.layer === 'large_model');

  // Build traditional layer result
  const traditional = {
    passed: acDetails.length === 0 || backend.pass,
    reason: acDetails.length > 0 ? `命中敏感词: ${acDetails.map(d => d.matched_text).join(', ')}` : undefined,
    keywords: acDetails.map(d => d.matched_text || '').filter(Boolean),
    score: acDetails.length > 0 ? acDetails[0].confidence : undefined
  };

  // Build small model layer result
  const smallModel = {
    passed: backend.pass,
    confidence: smallModelDetails.length > 0 ? smallModelDetails[0].confidence || 0 : 0,
    categories: smallModelDetails.map(d => d.category || '').filter(Boolean),
    reason: smallModelDetails.length > 0 ? backend.reason : '小模型未部署'
  };

  // Build large model layer result (optional)
  const largeModel = largeModelDetails.length > 0 ? {
    passed: backend.pass,
    confidence: largeModelDetails[0].confidence || 0,
    categories: largeModelDetails.map(d => d.category || '').filter(Boolean),
    analysis: backend.reason,
    reason: backend.reason
  } : undefined;

  // If no details but need_review, it means models are unavailable
  if (backend.details.length === 0 && backend.need_review) {
    smallModel.reason = backend.reason;
  }

  return {
    requestId: `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
    finalResult: backend.pass,
    traditional,
    smallModel,
    largeModel,
    processingTime: Date.now() - startTime
  };
}

/**
 * Check text content for moderation
 * @param data Text moderation request
 * @returns Moderation result with multi-layer checks
 */
export async function checkText(data: TextModerationRequest): Promise<TextModerationResponse> {
  const response = await request.post<BackendModerationResponse>(
    '/api/moderation/text/check',
    { content: data.content, content_type: data.scene },
    { timeout: 30000 }
  );
  return convertTextResponse(response);
}

/**
 * Check image content for moderation
 * @param data Image moderation request (base64 encoded)
 * @returns Moderation result with model checks
 */
export function checkImage(data: ImageModerationRequest) {
  return request.post<ImageModerationResponse>(
    '/api/moderation/image/check',
    data,
    { timeout: 60000 }
  );
}
