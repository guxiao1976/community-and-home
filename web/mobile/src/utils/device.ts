// Device ID utility for mobile H5
// Generates and persists a unique device identifier in localStorage

const DEVICE_ID_KEY = 'm_device_id';

let cachedDeviceId: string | null = null;

export function getDeviceId(): string {
  if (cachedDeviceId) return cachedDeviceId;

  let id = localStorage.getItem(DEVICE_ID_KEY);
  if (!id) {
    id = 'm-' + Date.now().toString(36) + '-' + Math.random().toString(36).substring(2, 10);
    localStorage.setItem(DEVICE_ID_KEY, id);
  }
  cachedDeviceId = id;
  return id;
}

export function getDeviceType(): string {
  // #ifdef H5
  return 'mobile-h5';
  // #endif
  // #ifdef MP-WEIXIN
  return 'mp-weixin';
  // #endif
  // #ifdef APP-PLUS
  return 'app';
  // #endif
}
