import { test, expect } from '@playwright/test';

test('stt engine dropdown is visible and saves to local storage', async ({ page }) => {
  await page.goto('/');
  const settingsBtn = page.getByTestId('settings-btn');
  await settingsBtn.click();

  const select = page.locator('#localSttEngine');
  await expect(select).toBeVisible();

  // Change to parakeet
  await select.selectOption('parakeet');

  // Use JS to click since it's hard to scroll in the custom overlay
  const saveBtn = page.getByTestId('settings-save-btn');
  await saveBtn.evaluate((el) => (el as HTMLElement).click());

  // Wait a little bit for local storage to save
  await page.waitForTimeout(100);

  const value = await page.evaluate(() => localStorage.getItem('local_stt_engine'));
  expect(value).toBe('parakeet');
});
