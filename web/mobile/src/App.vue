<script lang="ts">
import { isAuthenticated } from '@common/utils/auth';
import { getUserProfile } from '@/api/identity';
import { useUserStore } from '@/stores/user';

// 登录态修复：token 是权威，user 是 profile 缓存。
// 启动时若已登录（token 存在）但 user 未加载（登录后 profile 拉取失败 / tab 在登录前已挂载），
// 全局恢复 profile，覆盖所有页面，避免 isLoggedIn 误判未登录。
// 独立 <script> 块导出以便单测（App.spec.ts 覆盖 4 分支）。
export async function restoreUserProfile(): Promise<void> {
  if (!isAuthenticated()) return;
  const userStore = useUserStore();
  if (userStore.user) return;
  try {
    const user = await getUserProfile();
    userStore.setUser(user);
  } catch (err) {
    console.error('[App] 启动恢复用户资料失败', err);
  }
}
</script>

<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app';

onLaunch(() => {
  restoreUserProfile();
});

onShow(() => {
  // no-op
});

onHide(() => {
  // no-op
});
</script>

<style lang="scss">
/* Global app styles */
@use '@/uni.scss';

/* 单位体系：全站长度/字号使用 rem。根字号固定 16px（非响应式，不加 JS 动态缩放），
   rem 换算锚定 375px 设计稿：1rem = 16px；rpx→rem 按 1rpx=0.5px 折 16px 根字号（N rpx → N/32 rem） */
html {
  font-size: 16px;
}

page {
  font-family: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', Helvetica,
    'Segoe UI', Arial, 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei',
    sans-serif;
  font-size: $uni-font-size-base;
  color: $uni-text-color;
  background-color: $uni-bg-color;
  -webkit-font-smoothing: antialiased;
}
</style>
