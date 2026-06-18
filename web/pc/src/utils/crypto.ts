// RSA encryption utility using Web Crypto API
// Aligns with backend: common/pkg/crypto/rsa.go (RSA-OAEP + SHA-256 + Base64)

const PUBLIC_KEY_CACHE_KEY = 'rsa_public_key';

/**
 * Parse a PEM-formatted SPKI public key string into a CryptoKey.
 * Handles both "-----BEGIN PUBLIC KEY-----" and PKCS#1 formats.
 */
async function pemToCryptoKey(pem: string): Promise<CryptoKey> {
  // Remove PEM headers, footers, and whitespace
  const pemContent = pem
    .replace(/-----[A-Z ]+-----/g, '')
    .replace(/\s+/g, '');

  // Decode base64 to binary
  const binaryDer = Uint8Array.from(atob(pemContent), c => c.charCodeAt(0));

  return crypto.subtle.importKey(
    'spki',
    binaryDer.buffer,
    {
      name: 'RSA-OAEP',
      hash: 'SHA-256',
    },
    false, // not extractable
    ['encrypt']
  );
}

/**
 * Encrypt plaintext using RSA-OAEP with SHA-256 and return Base64-encoded ciphertext.
 *
 * Matches backend: common/pkg/crypto/rsa.go rsaEncryptWithPublicKey()
 * which uses rsa.EncryptOAEP with sha256.New() and base64.StdEncoding.
 */
export async function encryptWithPublicKey(
  plaintext: string,
  publicKeyPem: string
): Promise<string> {
  const publicKey = await pemToCryptoKey(publicKeyPem);

  const encoder = new TextEncoder();
  const plaintextBytes = encoder.encode(plaintext);

  const ciphertext = await crypto.subtle.encrypt(
    {
      name: 'RSA-OAEP',
    },
    publicKey,
    plaintextBytes
  );

  // Base64 encode (standard encoding, matches Go's base64.StdEncoding)
  return btoa(String.fromCharCode(...new Uint8Array(ciphertext)));
}

/**
 * Fetch the RSA public key from the auth service.
 * Result is cached in sessionStorage to avoid repeated requests.
 *
 * Requires backend: GET /api/auth/public-key returns { public_key: "PEM..." }
 */
export async function getPublicKey(): Promise<string> {
  // Return cached key if available
  const cached = sessionStorage.getItem(PUBLIC_KEY_CACHE_KEY);
  if (cached) {
    return cached;
  }

  // Dynamic import to avoid circular dependency
  const { default: request } = await import('@/utils/request');

  const data = await (request as any).get<{ public_key: string }>('/api/auth/public-key');

  const publicKey = data.public_key || data.publicKey;
  if (!publicKey) {
    throw new Error('Failed to get public key from server');
  }

  // Cache for this session
  sessionStorage.setItem(PUBLIC_KEY_CACHE_KEY, publicKey);
  return publicKey;
}

/**
 * Clear the cached public key (e.g., on logout).
 */
export function clearPublicKeyCache(): void {
  sessionStorage.removeItem(PUBLIC_KEY_CACHE_KEY);
}
