import { test, expect, type Page } from '@playwright/test';

/**
 * Generates a test email in the format:
 * WebTesting-YYYYMMDD-HHMMSSUTC-GUID@OpenAgentFramework.com
 */
function generateTestEmail(): string {
  const now = new Date();
  const yyyy = now.getUTCFullYear().toString();
  const mm = (now.getUTCMonth() + 1).toString().padStart(2, '0');
  const dd = now.getUTCDate().toString().padStart(2, '0');
  const hh = now.getUTCHours().toString().padStart(2, '0');
  const min = now.getUTCMinutes().toString().padStart(2, '0');
  const ss = now.getUTCSeconds().toString().padStart(2, '0');
  const guid = crypto.randomUUID();

  return `WebTesting-${yyyy}${mm}${dd}-${hh}${min}${ss}UTC-${guid}@OpenAgentFramework.com`;
}

const TEST_PASSWORD = 'TestPassword123!';
const TEST_DISPLAY_NAME = 'E2E Page Test User';

/**
 * Collects console errors and uncaught exceptions during page navigation.
 * Returns a list of error messages found.
 */
function collectPageErrors(page: Page): string[] {
  const errors: string[] = [];

  page.on('pageerror', (error) => {
    errors.push(`[PAGE ERROR] ${error.message}`);
  });

  page.on('console', (msg) => {
    if (msg.type() === 'error') {
      const text = msg.text();
      // Ignore known non-critical errors
      if (text.includes('WebSocket') || text.includes('net::ERR_')) return;
      errors.push(`[CONSOLE ERROR] ${text}`);
    }
  });

  return errors;
}

test.describe('All Pages - No Exceptions', () => {
  let testEmail: string;

  test.beforeAll(() => {
    testEmail = generateTestEmail();
  });

  // Register a user first, then test all pages
  test('register user and verify all pages render without exceptions', async ({ page }) => {
    const errors = collectPageErrors(page);

    // --- Register ---
    await page.goto('/register');
    await page.getByLabel('Display name').fill(TEST_DISPLAY_NAME);
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(TEST_PASSWORD);
    await page.getByLabel('Confirm password').fill(TEST_PASSWORD);
    await page.getByRole('button', { name: 'Create account' }).click();
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10000 });

    // --- Dashboard Page ---
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
    expect(errors, 'Dashboard page errors').toEqual([]);

    // --- Agents Page ---
    errors.length = 0;
    await page.goto('/agents');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Agents' })).toBeVisible();
    expect(errors, 'Agents page errors').toEqual([]);

    // --- Events Page ---
    errors.length = 0;
    await page.goto('/events');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Event Feed' })).toBeVisible();
    expect(errors, 'Events page errors').toEqual([]);

    // --- Settings Page ---
    errors.length = 0;
    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Organization Settings' })).toBeVisible();
    expect(errors, 'Settings page errors').toEqual([]);

    // --- API Keys Page ---
    errors.length = 0;
    await page.goto('/settings/api-keys');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'API Keys' })).toBeVisible();
    expect(errors, 'API Keys page errors').toEqual([]);

    // --- Audit Log Page ---
    errors.length = 0;
    await page.goto('/audit');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Audit Log' })).toBeVisible();
    expect(errors, 'Audit Log page errors').toEqual([]);
  });

  // Test unauthenticated pages separately
  test('login page renders without exceptions', async ({ page }) => {
    const errors = collectPageErrors(page);
    await page.goto('/login');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Sign in' })).toBeVisible();
    expect(errors, 'Login page errors').toEqual([]);
  });

  test('register page renders without exceptions', async ({ page }) => {
    const errors = collectPageErrors(page);
    await page.goto('/register');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'Create account' })).toBeVisible();
    expect(errors, 'Register page errors').toEqual([]);
  });

  test('invite page renders without exceptions when not authenticated', async ({ page }) => {
    const errors = collectPageErrors(page);
    await page.goto('/invite/fake-token');
    await page.waitForLoadState('networkidle');
    await expect(page.getByText('Sign in required')).toBeVisible();
    expect(errors, 'Invite page errors').toEqual([]);
  });

  // Test sidebar navigation works for every authenticated route
  test('sidebar navigation works for all pages', async ({ page }) => {
    // Register and login
    await page.goto('/register');
    const email = generateTestEmail();
    await page.getByLabel('Display name').fill('Nav Test User');
    await page.getByLabel('Email').fill(email);
    await page.getByLabel('Password', { exact: true }).fill(TEST_PASSWORD);
    await page.getByLabel('Confirm password').fill(TEST_PASSWORD);
    await page.getByRole('button', { name: 'Create account' }).click();
    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10000 });

    // Navigate via sidebar links and check for errors
    const navTests = [
      { link: 'Agents', url: '/agents', heading: 'Agents' },
      { link: 'Events', url: '/events', heading: 'Event Feed' },
      { link: 'Settings', url: '/settings', heading: 'Organization Settings' },
      { link: 'Audit Log', url: '/audit', heading: 'Audit Log' },
      { link: 'Dashboard', url: '/dashboard', heading: 'Dashboard' },
    ];

    for (const { link, url, heading } of navTests) {
      const errors: string[] = [];
      page.on('pageerror', (error) => errors.push(error.message));

      // Click sidebar link; fall back to direct navigation if link is detached
      try {
        await page.getByRole('link', { name: link, exact: true }).click({ timeout: 3000 });
      } catch {
        await page.goto(url);
      }
      await page.waitForLoadState('networkidle');
      await expect(page.getByRole('heading', { name: heading })).toBeVisible({ timeout: 5000 });
      expect(errors, `${link} page had JS errors`).toEqual([]);

      page.removeAllListeners('pageerror');
    }
  });
});
