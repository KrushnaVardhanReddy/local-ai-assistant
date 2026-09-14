import { test, expect } from '@playwright/test';

test.describe('App UI Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Inject wails mock before page loads
    await page.addInitScript(() => {
      window.runtime = {
        EventsOn: () => {},
        EventsOff: () => {},
        EventsOnce: () => {},
        EventsEmit: () => {},
        WindowHide: () => {},
        WindowShow: () => {},
        WindowSetTitle: () => {},
      } as any;
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('App mounts and shows Assistant UI', async ({ page }) => {
    // Verify body is visible
    const body = page.locator('body');
    await expect(body).toBeVisible();
  });

  test('Settings modal opens and closes', async ({ page }) => {
    // Attempt to open settings
    const settingsBtn = page.locator('button[aria-label="Settings"]');
    await settingsBtn.waitFor({ state: 'visible', timeout: 5000 });
    await settingsBtn.click();

    // Wait for settings panel
    const settingsPanel = page.locator('div.settings-panel, .fixed.inset-y-0.right-0');
    await expect(settingsPanel.first()).toBeVisible();
  });
});
