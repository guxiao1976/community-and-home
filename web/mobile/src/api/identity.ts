// Identity/Auth Service API — Uni-app mobile
import request from '@/utils/request';
import {
  getPublicKey as fetchPublicKey,
  encryptWithPublicKey,
  clearPublicKeyCache,
} from '@/utils/crypto';
import type {
  LoginRequest,
  LoginSmsRequest,
  RegisterRequest,
  LoginResponse,
  RefreshTokenRequest,
  RefreshTokenResponse,
  User,
} from '@common/types/identity';

// Re-export for convenience
export { fetchPublicKey, clearPublicKeyCache };

/**
 * Ensure public key is available. Fetches once and caches in sessionStorage.
 * Call this early (e.g., on login page mount) to preload the key.
 */
export async function ensurePublicKey(): Promise<string> {
  return fetchPublicKey();
}

/**
 * Encrypt the phone number with the cached RSA public key.
 */
async function encryptPhone(phone: string): Promise<string> {
  const publicKey = await fetchPublicKey();
  return encryptWithPublicKey(phone, publicKey);
}

/**
 * Login with phone and password.
 */
export async function login(
  phone: string,
  password: string,
  deviceId: string,
): Promise<LoginResponse> {
  const [encryptedPhone, encryptedPassword] = await Promise.all([
    encryptPhone(phone),
    encryptWithPublicKey(password, await fetchPublicKey()),
  ]);

  const data: LoginRequest = {
    encryptedPhone,
    encryptedPassword,
    deviceId,
    deviceType: 'mobile-h5',
  };

  const res = await request.post<LoginResponse>('/api/auth/login', data);
  return res as unknown as LoginResponse;
}

/**
 * Login with phone + SMS code (phone RSA-encrypted).
 */
export async function loginWithSms(
  phone: string,
  smsCode: string,
  deviceId: string,
): Promise<LoginResponse> {
  const encryptedPhone = await encryptPhone(phone);

  const data: LoginSmsRequest = {
    encryptedPhone,
    smsCode,
    deviceId,
    deviceType: 'mobile-h5',
  };

  const res = await request.post<LoginResponse>('/api/auth/login/sms', data);
  return res as unknown as LoginResponse;
}

/**
 * Register new user (phone RSA-encrypted).
 */
export async function register(params: {
  phone: string;
  password?: string;
  smsCode: string;
  nickname: string;
  deviceId: string;
}): Promise<LoginResponse> {
  const encryptedPhone = await encryptPhone(params.phone);

  let encryptedPassword: string | undefined;
  if (params.password) {
    encryptedPassword = await encryptWithPublicKey(
      params.password,
      await fetchPublicKey(),
    );
  }

  const data: RegisterRequest = {
    encryptedPhone,
    encryptedPassword,
    smsCode: params.smsCode,
    nickname: params.nickname,
    deviceId: params.deviceId,
    deviceType: 'mobile-h5',
  };

  const res = await request.post<LoginResponse>('/api/auth/register', data);
  return res as unknown as LoginResponse;
}

/**
 * Refresh access token.
 */
export async function refreshToken(
  token: string,
): Promise<RefreshTokenResponse> {
  const data: RefreshTokenRequest = { refreshToken: token };
  const res = await request.post<RefreshTokenResponse>(
    '/api/auth/token/refresh',
    data,
  );
  return res as unknown as RefreshTokenResponse;
}

/**
 * Get current user profile.
 */
export async function getUserProfile(): Promise<User> {
  const res = (await request.get<any>('/api/users/profile')) as unknown as any;
  // Backend wraps in { user: {...} }, extract the user object
  return (res?.user || res) as User;
}

/**
 * Send SMS verification code (phone in plain text — not encrypted).
 */
export async function sendSmsCode(phone: string): Promise<void> {
  await request.post('/api/auth/sms/send', { phone });
}

/**
 * Logout — revoke the current device session. Requires JWT.
 * kickAllDevices=true 时同时撤销该用户所有设备会话（本机登出不影响其他设备）。
 * 后端接口已存在：POST /api/auth/logout（401 时拦截器会处理 token 刷新/失效）。
 */
export async function logout(deviceId: string, kickAllDevices = false): Promise<void> {
  await request.post('/api/auth/logout', {
    deviceId,
    kickAllDevices,
  });
}
