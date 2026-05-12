import request from '@/utils/request';
import type {
  TextModerationRequest,
  TextModerationResponse,
  ImageModerationRequest,
  ImageModerationResponse
} from '@common/types/moderation';

/**
 * Check text content for moderation
 * @param data Text moderation request
 * @returns Moderation result with multi-layer checks
 */
export function checkText(data: TextModerationRequest) {
  return request.post<TextModerationResponse>(
    '/api/moderation/text/check',
    data,
    { timeout: 30000 }
  );
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
