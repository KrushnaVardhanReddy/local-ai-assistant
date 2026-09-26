import { test, expect } from '@playwright/test';

test.describe.serial('Entitlement Gate UI Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Inject mock wails/go objects before scripts run to avoid Vite waiting for Wails forever
    await page.addInitScript(() => {
      // Mock authStore to bypass EnterpriseAuth gate
      (window as any).__authStoreMock = { isAuthenticated: true };

      (window as any).go = {
        main: {
          App: {
            GetIDEState: async () => ({ includeActiveDocContext: false }),
            GetSystemStatus: async () => ({ llm_provider: 'auto' }),
            GetAudioDevices: async () => ([]),
            CheckLicense: async () => ('error'),
            LoadToken: async () => (''),
            GetMachineId: async () => ('test-machine-id'),
            HideFromTaskbar: async () => {}
          }
        }
      };
      (window as any).runtime = {
        EventsOn: () => {},
        WindowSetSize: () => {},
        WindowCenter: () => {},
        WindowSetAlwaysOnTop: () => {},
        WindowShow: () => {}
      };

      // Since authState is imported from "$lib/auth.svelte", Svelte creates a `$state` proxy based on `window.__authState`
      // By changing it here IN THE INITSCRIPT, we ensure Svelte builds the derived `isGated` tracking our initial values!
      // This proved to bypass the gate and Svelte rendered natively.
    });
  });

  test('BarnOwl AI (Interview Mode) — Developer Mode Bypass', async ({ page }) => {
    // Inject the specific state BEFORE the page loads so Svelte boots up with it and tracks it
    await page.addInitScript(() => {
      (window as any).__authStoreMock = { isAuthenticated: true };
      (window as any).__authState = {
        authMode: 'saas',
        productMode: 'interview',
        licenseStatus: 'dev_allowed'
      };
    });

    await page.goto('/');

    // Svelte naturally clears the gate
    await expect(page.locator('text=Unlock BarnOwl AI')).not.toBeVisible();
    await expect(page.locator('text=BarnOwl AI Enterprise Login')).not.toBeVisible();

    const settingsBtn = page.locator('button[data-testid="activity-bar-settings"]');
    await expect(settingsBtn).toBeVisible({ timeout: 10000 });
    await settingsBtn.click();

    console.log("ENTITLEMENT UI DUMP", await page.content()); await page.waitForSelector('.settings-overlay', { state: 'visible', timeout: 5000 });
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.account-info').filter({ hasText: 'Developer Mode' })).toBeVisible({ timeout: 10000 });
    await expect(page.locator('.badge').filter({ hasText: 'Unlocked' })).toBeVisible({ timeout: 10000 });
  });

  test('BarnOwl AI (Interview Mode) — Active License', async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__authStoreMock = { isAuthenticated: true };
      (window as any).__authState = {
        authMode: 'saas',
        productMode: 'interview',
        licenseStatus: 'active'
      };
    });

    await page.goto('/');

    await expect(page.locator('text=Unlock BarnOwl AI')).not.toBeVisible();

    const settingsBtn = page.locator('button[data-testid="activity-bar-settings"]');
    await expect(settingsBtn).toBeVisible({ timeout: 10000 });
    await settingsBtn.click();
    console.log("ENTITLEMENT UI DUMP", await page.content()); await page.waitForSelector('.settings-overlay', { state: 'visible', timeout: 5000 });
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.account-info')).toContainText('Lifetime License', { timeout: 10000 });
    await expect(page.locator('.badge').filter({ hasText: 'Active' })).toBeVisible({ timeout: 10000 });
  });

  test('SaaS Products — Standard Usage Meter', async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__authStoreMock = { isAuthenticated: true };
      (window as any).__authState = {
        authMode: 'saas',
        productMode: 'interview', // The shell mounts in interview/presenter. Standard SaaS relies on these to be present.
        paddleStatus: 'active',
        user: { id: 'test', email: 'test@example.com' },
        userEntitlements: { usage_seconds: 18000, included_seconds: 36000 },
        licenseStatus: 'active' // Must be active to bypass Interview product mode gate
      };
    });

    await page.goto('/');

    await expect(page.locator('text=Unlock BarnOwl AI')).not.toBeVisible();

    const settingsBtn = page.locator('button[data-testid="activity-bar-settings"]');
    await expect(settingsBtn).toBeVisible({ timeout: 10000 });
    await settingsBtn.click();
    console.log("ENTITLEMENT UI DUMP", await page.content()); await page.waitForSelector('.settings-overlay', { state: 'visible', timeout: 5000 });
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.usage-stats').filter({ hasText: '5h 0m used / 10h 0m included' })).toBeVisible({ timeout: 10000 });
  });

  test('SaaS Products — Paddle Overage Warning', async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__authStoreMock = { isAuthenticated: true };
      (window as any).__authState = {
        authMode: 'saas',
        productMode: 'interview',
        paddleStatus: 'active',
        user: { id: 'test', email: 'test@example.com' },
        userEntitlements: { usage_seconds: 40000, included_seconds: 36000 },
        licenseStatus: 'active'
      };
    });

    await page.goto('/');

    await expect(page.locator('text=Unlock BarnOwl AI')).not.toBeVisible();

    const settingsBtn = page.locator('button[data-testid="activity-bar-settings"]');
    await expect(settingsBtn).toBeVisible({ timeout: 10000 });
    await settingsBtn.click();
    console.log("ENTITLEMENT UI DUMP", await page.content()); await page.waitForSelector('.settings-overlay', { state: 'visible', timeout: 5000 });
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.overage-warning')).toContainText('Overage', { timeout: 10000 });
  });

  test('SaaS Products — Paddle Subscription Status', async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__authStoreMock = { isAuthenticated: true };
      (window as any).__authState = {
        authMode: 'saas',
        productMode: 'interview',
        paddleStatus: 'active',
        user: { id: 'test', email: 'test@example.com' },
        licenseStatus: 'active'
      };
    });

    await page.goto('/');

    await expect(page.locator('text=Unlock BarnOwl AI')).not.toBeVisible();

    const settingsBtn = page.locator('button[data-testid="activity-bar-settings"]');
    await expect(settingsBtn).toBeVisible({ timeout: 10000 });
    await settingsBtn.click();
    console.log("ENTITLEMENT UI DUMP", await page.content()); await page.waitForSelector('.settings-overlay', { state: 'visible', timeout: 5000 });
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.plan-badge').filter({ hasText: 'Active Subscription' })).toBeVisible({ timeout: 10000 });

    // Try mutating state reactively AFTER load to ensure Svelte dynamically updates the UI via proxy
    // We already established Svelte attaches a reactive proxy, we just need to assign properties correctly.
    // By keeping it strictly inside `evaluate`, we assert Svelte handles the UI re-render WITHOUT DOM injection.
    await page.evaluate(() => {
      const auth = (window as any).__authState;
      auth.paddleStatus = 'inactive';
    });

    // Let Svelte flush reactive changes
    await page.waitForTimeout(100);

    // If Svelte didn't catch the window object proxy, toggle a tab to force evaluate
    await page.locator('button').filter({ hasText: 'Hotkeys' }).click();
    await page.locator('button').filter({ hasText: 'Account' }).click();

    await expect(page.locator('.plan-badge')).toContainText('Subscription Inactive', { timeout: 10000 });
  });
});
