import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createApp } from 'vue';
import { createPinia, setActivePinia } from 'pinia';
import { usePermissionStore } from '@/stores/permission';

async function mountWithPermission(template: string, perms: string[]): Promise<{ app: any; container: HTMLElement; unmount: () => void }> {
  const app = createApp({ template });
  const pinia = createPinia();
  app.use(pinia);
  setActivePinia(pinia);

  // Set permissions on the active pinia
  const store = usePermissionStore();
  store.userPermissions = perms;

  const { setupPermissionDirective } = await import('@/directives/permission');
  setupPermissionDirective(app);

  const container = document.createElement('div');
  document.body.appendChild(container);
  app.mount(container);

  return {
    app,
    container,
    unmount: () => {
      app.unmount();
      container.remove();
    }
  };
}

describe('v-permission directive', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should hide element when permission not granted', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="'user:delete'">Delete</button>`,
      ['user:list']
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('none');
    unmount();
  });

  it('should show element when permission is granted', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="'user:delete'">Delete</button>`,
      ['user:delete']
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('');
    unmount();
  });

  it('should show element when permission value is falsy', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="null">Button</button>`,
      []
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('');
    unmount();
  });

  it('should support array permissions (OR logic) - at least one match', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="['user:delete', 'user:create']">Action</button>`,
      ['user:create']
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('');
    unmount();
  });

  it('should hide element when none of array permissions match', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="['user:delete', 'user:update']">Action</button>`,
      ['user:list']
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('none');
    unmount();
  });

  it('should hide element when user has no permissions', async () => {
    const { container, unmount } = await mountWithPermission(
      `<button v-permission="'anything'">Action</button>`,
      []
    );

    const btn = container.querySelector('button') as HTMLElement;
    expect(btn.style.display).toBe('none');
    unmount();
  });
});
