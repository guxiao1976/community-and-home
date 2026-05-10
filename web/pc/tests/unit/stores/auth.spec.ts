import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setActivePinia, createPinia } from 'pinia';
import { useAuthStore } from '@/stores/auth';
import * as identityApi from '@/api/identity';
import * as authUtils from '@common/utils/auth';

// Mock the API
vi.mock('@/api/identity');
vi.mock('@common/utils/auth');

describe('Auth Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.clearAllMocks();

    // Mock auth utils
    vi.mocked(authUtils.getAccessToken).mockReturnValue(null);
    vi.mocked(authUtils.getRefreshToken).mockReturnValue(null);
    vi.mocked(authUtils.getTokenExpiry).mockReturnValue(0);
    vi.mocked(authUtils.setTokens).mockImplementation(() => {});
    vi.mocked(authUtils.clearTokens).mockImplementation(() => {});
  });

  it('should initialize with default state', () => {
    const store = useAuthStore();

    expect(store.isAuthenticated).toBe(false);
    expect(store.user).toBeNull();
    expect(store.accessToken).toBeNull();
    expect(store.refreshToken).toBeNull();
  });

  it('should login successfully with password', async () => {
    const store = useAuthStore();
    const mockResponse = {
      user: {
        id: 1,
        phone: '13800000000',
        nickname: 'Test User',
        avatar: '',
        userType: 1,
        status: 1,
        verificationStatus: 1,
        scope: '',
        lastLoginAt: '',
        createdAt: '',
        updatedAt: '',
        deleteTime: 0
      },
      accessToken: 'mock-access-token',
      refreshToken: 'mock-refresh-token',
      expiresIn: 86400
    };

    vi.mocked(identityApi.login).mockResolvedValue(mockResponse);

    await store.login('13800000000', 'Admin@123456');

    expect(store.isAuthenticated).toBe(true);
    expect(store.user).toEqual(mockResponse.user);
    expect(store.accessToken).toBe('mock-access-token');
    expect(store.refreshToken).toBe('mock-refresh-token');
    expect(authUtils.setTokens).toHaveBeenCalledWith(
      'mock-access-token',
      'mock-refresh-token',
      86400
    );
  });

  it('should logout successfully', async () => {
    const store = useAuthStore();

    // Set initial state
    store.user = {
      id: 1,
      phone: '13800000000',
      nickname: 'Test User',
      avatar: '',
      userType: 1,
      status: 1,
      verificationStatus: 1,
      scope: '',
      lastLoginAt: '',
      createdAt: '',
      updatedAt: '',
      deleteTime: 0
    };
    store.accessToken = 'mock-access-token';
    store.refreshToken = 'mock-refresh-token';

    vi.mocked(identityApi.logout).mockResolvedValue(undefined);

    await store.logout();

    expect(store.isAuthenticated).toBe(false);
    expect(store.user).toBeNull();
    expect(store.accessToken).toBeNull();
    expect(store.refreshToken).toBeNull();
    expect(authUtils.clearTokens).toHaveBeenCalled();
  });

  it('should refresh token successfully', async () => {
    const store = useAuthStore();
    store.refreshToken = 'old-refresh-token';

    const mockResponse = {
      accessToken: 'new-access-token',
      refreshToken: 'new-refresh-token',
      expiresIn: 86400
    };

    vi.mocked(identityApi.refreshToken).mockResolvedValue(mockResponse);

    await store.refreshAccessToken();

    expect(store.accessToken).toBe('new-access-token');
    expect(store.refreshToken).toBe('new-refresh-token');
    expect(authUtils.setTokens).toHaveBeenCalledWith(
      'new-access-token',
      'new-refresh-token',
      86400
    );
  });

  it('should restore session from storage', () => {
    vi.mocked(authUtils.getAccessToken).mockReturnValue('stored-access-token');
    vi.mocked(authUtils.getRefreshToken).mockReturnValue('stored-refresh-token');
    vi.mocked(authUtils.getTokenExpiry).mockReturnValue(Date.now() + 86400000);

    const store = useAuthStore();
    store.restoreSession();

    expect(store.isAuthenticated).toBe(true);
    expect(store.accessToken).toBe('stored-access-token');
    expect(store.refreshToken).toBe('stored-refresh-token');
  });

  it('should handle login failure', async () => {
    const store = useAuthStore();

    vi.mocked(identityApi.login).mockRejectedValue(
      new Error('Invalid credentials')
    );

    await expect(
      store.login('13800000000', 'WrongPassword')
    ).rejects.toThrow('Invalid credentials');

    expect(store.isAuthenticated).toBe(false);
    expect(store.user).toBeNull();
  });

  it('should handle refresh token failure', async () => {
    const store = useAuthStore();
    store.refreshToken = 'expired-refresh-token';

    vi.mocked(identityApi.refreshToken).mockRejectedValue(
      new Error('Refresh token expired')
    );

    await expect(store.refreshAccessToken()).rejects.toThrow('Refresh token expired');
  });

  it('should check if token is expiring', () => {
    const store = useAuthStore();

    // Token expiring in 3 minutes
    store.tokenExpiry = Date.now() + 3 * 60 * 1000;
    expect(store.isTokenExpiring).toBe(true);

    // Token expiring in 10 minutes
    store.tokenExpiry = Date.now() + 10 * 60 * 1000;
    expect(store.isTokenExpiring).toBe(false);

    // No token
    store.tokenExpiry = 0;
    expect(store.isTokenExpiring).toBe(false);
  });
});
