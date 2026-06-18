// Device identification for multi-device login management

const DEVICE_ID_KEY = 'device_id';

/**
 * Generate a UUID v4 using crypto.getRandomValues().
 * Works in all modern browsers (unlike crypto.randomUUID() which is newer).
 */
function generateUUID(): string {
  // Prefer crypto.randomUUID() if available (Chrome 92+, Firefox 95+, Safari 15.4+)
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID();
  }

  // Fallback: manual UUID v4 using crypto.getRandomValues()
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (crypto.getRandomValues(new Uint8Array(1))[0] & 15) >> (c === 'x' ? 0 : 3);
    return (c === 'x' ? r : (r & 0x3) | 0x8).toString(16);
  });
}

/**
 * Get or create a persistent device identifier.
 * Uses crypto.getRandomValues() for generation and localStorage for persistence.
 * This maps to the device_id field in LoginRequest proto.
 */
export function getDeviceId(): string {
  let deviceId = localStorage.getItem(DEVICE_ID_KEY);
  if (!deviceId) {
    deviceId = generateUUID();
    localStorage.setItem(DEVICE_ID_KEY, deviceId);
  }
  return deviceId;
}

/**
 * Get the device type string for the current platform.
 */
export function getDeviceType(): string {
  return 'web';
}
