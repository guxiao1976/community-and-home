<script setup lang="ts">
import { onLaunch, onShow, onHide } from '@dcloudio/uni-app';

onLaunch(() => {
  console.log('[App] Launch — 社区家园 mobile');
});

onShow(() => {
  console.log('[App] Show');
});

onHide(() => {
  console.log('[App] Hide');
});
</script>

<style lang="scss">
/* Global app styles */
@use '@/uni.scss';

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
