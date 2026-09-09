import { test, expect } from '@playwright/test';

test.describe('Mobile Layout Tests', () => {
  test('main app layout and bottom buttons are visible', async ({ page, isMobile }) => {
    test.skip(!isMobile, 'Only run on mobile devices');

    await page.goto('/');

    // Ensure main app shell fits within 375px width
    const appShell = page.locator('.app-shell');
    await expect(appShell).toBeVisible();

    // Check width of the body or main container to not exceed 375px
    const bodyBox = await page.locator('body').boundingBox();
    expect(bodyBox?.width).toBeLessThanOrEqual(375);

    // Assert that the element is visible on the page (exists in DOM and is rendered)
    const stealthInput = page.getByPlaceholder(/Silent chat/i);
    await expect(stealthInput).toBeVisible();

    const micButton = page.locator('#mock-mode-btn');
    await expect(micButton).toBeVisible();

    // Settings gear is visible
    const settingsBtn = page.getByTestId('settings-btn');
    await expect(settingsBtn).toBeVisible();

    // The settings modal is still usable and scrollable on small screens
    await settingsBtn.click({ force: true });
    const settingsPanel = page.locator('.settings-panel');
    await expect(settingsPanel).toBeVisible();

    const accountHeader = page.locator('h2', { hasText: 'Account' });
    await expect(accountHeader).toBeVisible();
  });
});
