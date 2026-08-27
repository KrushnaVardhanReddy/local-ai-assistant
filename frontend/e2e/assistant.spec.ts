import { test, expect } from '@playwright/test';

test('assistant overlay renders', async ({ page }) => {
  await page.goto('/');
  const panel = page.getByTestId('assistant-panel');
  await expect(panel).toBeVisible();
});

test('mic status dot renders', async ({ page }) => {
  await page.goto('/');
  const micStatus = page.getByTestId('mic-status');
  await expect(micStatus).toBeVisible();
});

test('settings gear is clickable', async ({ page }) => {
  await page.goto('/');
  const settingsBtn = page.getByTestId('settings-btn');
  await settingsBtn.click();
  const urlInput = page.getByTestId('backend-url-input');
  await expect(urlInput).toBeVisible();
});

test('dev mode toggle exists', async ({ page }) => {
  await page.goto('/');
  const settingsBtn = page.getByTestId('settings-btn');
  await settingsBtn.click();
  const devModeToggle = page.getByTestId('dev-mode-toggle');
  await expect(devModeToggle).toBeVisible();
});
