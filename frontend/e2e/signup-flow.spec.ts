import { test, expect } from '@playwright/test';

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
const TEST_DISPLAY_NAME = 'Web Test User';

test.describe('Signup Flow', () => {
  test('should navigate to registration page', async ({ page }) => {
    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Create account' })).toBeVisible();
  });

  test('should complete full signup and reach dashboard', async ({ page }) => {
    const testEmail = generateTestEmail();

    await page.goto('/register');
    await expect(page.getByRole('heading', { name: 'Create account' })).toBeVisible();

    await page.getByLabel('Display name').fill(TEST_DISPLAY_NAME);
    await page.getByLabel('Email').fill(testEmail);
    await page.getByLabel('Password', { exact: true }).fill(TEST_PASSWORD);
    await page.getByLabel('Confirm password').fill(TEST_PASSWORD);

    await page.getByRole('button', { name: 'Create account' }).click();

    await expect(page).toHaveURL(/\/dashboard/, { timeout: 10000 });
    await expect(page.getByRole('heading', { name: 'Dashboard' })).toBeVisible();
  });

  test('should show validation errors for invalid input', async ({ page }) => {
    await page.goto('/register');
    await page.getByRole('button', { name: 'Create account' }).click();

    await expect(page.getByText('Name must be at least 2 characters')).toBeVisible();
    await expect(page.getByText('Invalid email address')).toBeVisible();
  });

  test('should show error for password mismatch', async ({ page }) => {
    await page.goto('/register');

    await page.getByLabel('Display name').fill(TEST_DISPLAY_NAME);
    await page.getByLabel('Email').fill(generateTestEmail());
    await page.getByLabel('Password', { exact: true }).fill(TEST_PASSWORD);
    await page.getByLabel('Confirm password').fill('DifferentPassword123!');
    await page.getByRole('button', { name: 'Create account' }).click();

    await expect(page.getByText('Passwords do not match')).toBeVisible();
  });

  test('should navigate from login to register', async ({ page }) => {
    await page.goto('/login');
    await page.getByText('Sign up').click();
    await expect(page).toHaveURL(/\/register/);
  });
});
