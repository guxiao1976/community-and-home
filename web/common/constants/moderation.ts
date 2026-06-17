// source_type display name mapping
export const SOURCE_TYPE_LABELS: Record<string, string> = {
  notice: '通知公告',
  lost_found: '寻失互助',
  certification: '房主认证',
  nickname: '用户昵称',
};

// moderation_status display mapping
export const MODERATION_STATUS_MAP: Record<number, { label: string; type: string }> = {
  0: { label: '待审核', type: 'info' },
  1: { label: '机器通过', type: 'success' },
  2: { label: '机器不通过', type: 'danger' },
  3: { label: '人审通过', type: '' },
  4: { label: '人审不通过', type: 'danger' },
};
