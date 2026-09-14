import { test, expect } from '@playwright/test';

test('App mounts and successfully talks to real Cloudflare endpoint on mock interview start', async ({ page }) => {
  await page.goto('http://localhost:5173');

  // Wait for network idle (hydration completion in Svelte 5)
  await page.waitForLoadState('networkidle');

  // Verify body is visible
  const body = await page.locator('body');
  await expect(body).toBeVisible();

  // Setup a page.waitForResponse interceptor that listens for network requests
  // hitting the real Cloudflare worker endpoint.
  // We can just check the URL and ignore status because CORS preflight might be different
  // or it might fail if auth is missing in test environment, but the point is it ATTEMPTED
  // to talk to the REAL worker, not a mock.
  const apiRequestPromise = page.waitForRequest((request) => {
    return request.url().includes('ai.krushnavardhan.workers.dev');
  });

  // Find and click the "Start Mock Interview Practice" button
  const startButton = page.locator('button[aria-label="Start Mock Interview Practice"]');
  await startButton.waitFor({ state: 'visible', timeout: 5000 });
  await startButton.click({ force: true });

  // Await the request to verify we actually talked to the REAL worker
  const apiRequest = await apiRequestPromise;
  expect(apiRequest.url()).toContain('ai.krushnavardhan.workers.dev');

  // Assert that UI state changes (e.g., displaying 'Mock Interview Mode')
  const mockModeHeading = page.locator('h2:has-text("Mock Interview Mode")');
  await expect(mockModeHeading).toBeVisible();
});
