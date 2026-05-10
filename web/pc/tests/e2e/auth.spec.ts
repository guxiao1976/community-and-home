import { test, expect } from '@playwright/test';

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Clear storage before each test
    await page.goto('/');
    await page.evaluate(() => {
      localStorage.clear();
      sessionStorage.clear();
    });
  });

  test('should display login page', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('h1')).toContainText('社区家园管理平台');
    await expect(page.locator('text=密码登录')).toBeVisible();
    await expect(page.locator('text=短信登录')).toBeVisible();
  });

  test('should login with phone and password', async ({ page }) => {
    await page.goto('/login');

    // Fill in login form
    await page.fill('input[placeholder="请输入手机号"]', '13800000000');
    await page.fill('input[placeholder="请输入密码"]', 'Admin@123456');

    // Click login button
    await page.click('button:has-text("登录")');

    // Should redirect to dashboard
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('h2')).toContainText('欢迎使用社区家园管理平台');
  });

  test('should show error for invalid credentials', async ({ page }) => {
    await page.goto('/login');

    // Fill in invalid credentials
    await page.fill('input[placeholder="请输入手机号"]', '13800000000');
    await page.fill('input[placeholder="请输入密码"]', 'WrongPassword123!');

    // Click login button
    await page.click('button:has-text("登录")');

    // Should show error message
    await expect(page.locator('.el-message--error')).toBeVisible();
  });

  test('should validate phone number format', async ({ page }) => {
    await page.goto('/login');

    // Fill in invalid phone number
    await page.fill('input[placeholder="请输入手机号"]', '12345');
    await page.fill('input[placeholder="请输入密码"]', 'Admin@123456');

    // Click login button
    await page.click('button:has-text("登录")');

    // Should show validation error
    await expect(page.locator('.el-form-item__error')).toBeVisible();
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[placeholder="请输入手机号"]', '13800000000');
    await page.fill('input[placeholder="请输入密码"]', 'Admin@123456');
    await page.click('button:has-text("登录")');

    // Wait for dashboard
    await expect(page).toHaveURL('/dashboard');

    // Click logout button
    await page.click('button:has-text("退出登录")');

    // Confirm logout
    await page.click('button:has-text("确定")');

    // Should redirect to login page
    await expect(page).toHaveURL('/login');
  });

  test('should redirect to login when accessing protected route without auth', async ({ page }) => {
    await page.goto('/dashboard');

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/);
  });

  test('should restore session after page reload', async ({ page }) => {
    // Login first
    await page.goto('/login');
    await page.fill('input[placeholder="请输入手机号"]', '13800000000');
    await page.fill('input[placeholder="请输入密码"]', 'Admin@123456');
    await page.click('button:has-text("登录")');

    // Wait for dashboard
    await expect(page).toHaveURL('/dashboard');

    // Reload page
    await page.reload();

    // Should still be on dashboard (session restored)
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('h2')).toContainText('欢迎使用社区家园管理平台');
  });

  test('should display register page', async ({ page }) => {
    await page.goto('/register');

    await expect(page.locator('h1')).toContainText('用户注册');
    await expect(page.locator('input[placeholder="请输入手机号"]')).toBeVisible();
    await expect(page.locator('input[placeholder="请输入验证码"]')).toBeVisible();
    await expect(page.locator('input[placeholder*="昵称"]')).toBeVisible();
  });

  test('should navigate between login and register', async ({ page }) => {
    await page.goto('/login');

    // Click register link
    await page.click('text=还没有账号？立即注册');
    await expect(page).toHaveURL('/register');

    // Click login link
    await page.click('text=已有账号？立即登录');
    await expect(page).toHaveURL('/login');
  });
});
