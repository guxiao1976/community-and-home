import { createApp } from 'vue';
import { createPinia } from 'pinia';
import piniaPluginPersistedstate from 'pinia-plugin-persistedstate';
import ElementPlus from 'element-plus';
import 'element-plus/dist/index.css';
import * as ElementPlusIconsVue from '@element-plus/icons-vue';
import zhCn from 'element-plus/es/locale/lang/zh-cn';

import App from './App.vue';
import router from './router';
import './router/guards';
import './styles/global.scss';
import { setupPermissionDirective } from '@/directives/permission';

const app = createApp(App);

// Pinia
const pinia = createPinia();
pinia.use(piniaPluginPersistedstate);
app.use(pinia);

// Vue Router
app.use(router);

// Element Plus
app.use(ElementPlus, {
  locale: zhCn
});

// Register Element Plus icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component);
}

// Register permission directive
setupPermissionDirective(app);

app.mount('#app');
