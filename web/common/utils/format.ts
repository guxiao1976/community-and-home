// Data formatting utilities

import dayjs from 'dayjs';

/**
 * Format date to YYYY-MM-DD HH:mm:ss
 */
export function formatDateTime(date: string | Date | number): string {
  if (!date) return '';
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss');
}

/**
 * Format date to YYYY-MM-DD
 */
export function formatDate(date: string | Date | number): string {
  if (!date) return '';
  return dayjs(date).format('YYYY-MM-DD');
}

/**
 * Format time to HH:mm:ss
 */
export function formatTime(date: string | Date | number): string {
  if (!date) return '';
  return dayjs(date).format('HH:mm:ss');
}

/**
 * Format relative time (e.g., "2 hours ago")
 */
export function formatRelativeTime(date: string | Date | number): string {
  if (!date) return '';
  const now = dayjs();
  const target = dayjs(date);
  const diffMinutes = now.diff(target, 'minute');
  const diffHours = now.diff(target, 'hour');
  const diffDays = now.diff(target, 'day');

  if (diffMinutes < 1) return '刚刚';
  if (diffMinutes < 60) return `${diffMinutes}分钟前`;
  if (diffHours < 24) return `${diffHours}小时前`;
  if (diffDays < 7) return `${diffDays}天前`;
  return formatDate(date);
}

/**
 * Format number with thousand separators
 */
export function formatNumber(num: number): string {
  if (num === null || num === undefined) return '';
  return num.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

/**
 * Format file size (bytes to KB/MB/GB)
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
}

/**
 * Format percentage
 */
export function formatPercentage(value: number, total: number, decimals: number = 2): string {
  if (total === 0) return '0%';
  return ((value / total) * 100).toFixed(decimals) + '%';
}
