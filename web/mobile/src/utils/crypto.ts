// RSA encryption utility using Web Crypto API
// Aligns with backend: common/pkg/crypto/rsa.go (RSA-OAEP + SHA-256 + Base64)
//
// Requirement: Web Crypto API is available in HTTPS and localhost contexts.
// Uni-app H5 dev server (localhost:3004) satisfies this requirement.

const PUBLIC_KEY_CACHE_KEY = 'rsa_public_key';

/**
 * Parse a PEM-formatted SPKI public key string into a CryptoKey.
 */
async function pemToCryptoKey(pem: string): Promise<CryptoKey> {
  const pemContent = pem
    .replace(/-----[A-Z ]+-----/g, '')
    .replace(/\s+/g, '');

  const binaryDer = Uint8Array.from(atob(pemContent), (c) => c.charCodeAt(0));

  return crypto.subtle.importKey(
    'spki',
    binaryDer.buffer,
    { name: 'RSA-OAEP', hash: 'SHA-256' },
    false,
    ['encrypt'],
  );
}

/**
 * Encrypt plaintext using RSA-OAEP with SHA-256 and return Base64-encoded ciphertext.
 *
 * Matches backend: common/pkg/crypto/rsa.go rsaEncryptWithPublicKey()
 * which uses rsa.EncryptOAEP(sha256.New(), ...) + base64.StdEncoding.
 */
export async function encryptWithPublicKey(
  plaintext: string,
  publicKeyPem: string,
): Promise<string> {
  const publicKey = await pemToCryptoKey(publicKeyPem);
  const encoder = new TextEncoder();
  const plaintextBytes = encoder.encode(plaintext);
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'RSA-OAEP' },
    publicKey,
    plaintextBytes,
  );
  return btoa(String.fromCharCode(...new Uint8Array(ciphertext)));
}

/**
 * Fetch the RSA public key from the auth service.
 * Result is cached in sessionStorage to avoid repeated requests.
 *
 * Endpoint: GET /api/auth/public-key → { public_key: "PEM..." }
 */
export async function getPublicKey(): Promise<string> {
  const cached = sessionStorage.getItem(PUBLIC_KEY_CACHE_KEY);
  if (cached) {
    console.log('[Crypto] Using cached public key');
    return cached;
  }

  const { default: request } = await import('@/utils/request');

  console.log('[Crypto] Fetching public key from /api/auth/public-key...');
  const raw = await (request as any).get('/api/auth/public-key');
  console.log('[Crypto] Raw response:', typeof raw, JSON.stringify(raw).substring(0, 200));

  // Backend may return public_key or publicKey
  const publicKey = raw?.public_key || raw?.publicKey;
  if (!publicKey) {
    console.error('[Crypto] Response missing public_key field:', raw);
    throw new Error('Failed to get public key from server');
  }

  sessionStorage.setItem(PUBLIC_KEY_CACHE_KEY, publicKey);
  return publicKey;
}

/**
 * Clear the cached public key (e.g., on logout).
 */
export function clearPublicKeyCache(): void {
  sessionStorage.removeItem(PUBLIC_KEY_CACHE_KEY);
}
