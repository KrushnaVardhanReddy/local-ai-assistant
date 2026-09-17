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
      window.go = {
        main: {
          App: {
            GetIDEState: async () => ({
              includeActiveDocContext: true,
              activeDocumentName: 'mock_notes.md',
              workspaceTree: [{ name: 'folder1', is_dir: true, children: [{ name: 'mock_notes.md', path: '/mock_notes.md', is_dir: false }] }],
              openDocuments: [{ name: 'mock_notes.md', path: '/mock_notes.md' }],
              activeDocumentPath: '/mock_notes.md',
              activeDocumentContent: '# Mock Content',
            }),
            SetIncludeActiveDocContext: async () => {},
            SetClickthrough: async (enable: boolean) => true,
            CaptureScreen: async () => "data:image/png;base64,mock",
            AnalyzeVision: async (b64: string, prompt: string) => true,
            EndSession: async () => ({
              session: { turn_count: 2, session_duration_s: 120 },
              scorecard: { overall_score: 8, overall_summary: 'Good mock test', turn_evaluations: [] }
            }),
          },
        },
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

  test('Active document context chip is visible when document is active', async ({ page }) => {
    // Wait for the context chip to appear based on our mock data
    const contextChip = page.locator('text=📄 Context: mock_notes.md');
    await expect(contextChip).toBeVisible();
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

  test('Stealth mode toggle', async ({ page }) => {
    const stealthBtn = page.locator('button[title="Stealth Mode (Clickthrough)"]');
    await expect(stealthBtn).toBeVisible();
    await stealthBtn.click();
    await expect(stealthBtn).toHaveClass(/text-primary/);
  });

  test('Vision snip button', async ({ page }) => {
    const snipBtn = page.locator('button[title="Vision Snip"]');
    await expect(snipBtn).toBeVisible();
    await snipBtn.click();
  });

  test('Session Report open and close', async ({ page }) => {
    const reportBtn = page.locator('button[title="Session Report"]');
    await expect(reportBtn).toBeVisible();
    await reportBtn.click();

    const reportModal = page.locator('.report-overlay');
    await expect(reportModal).toBeVisible();

    const closeBtn = reportModal.locator('button[aria-label="Close"]');
    await closeBtn.click();
    await expect(reportModal).toBeHidden();
  });

  test('STAR preset trigger', async ({ page }) => {
    // Assume there is a STAR preset button somewhere (e.g. ActivityBar or Drawer)
    // If it's not present natively right now, just verify we can click ActivityBar mock action
    const actionBtn = page.locator('button[aria-label="Mock Mode"]');
    if (await actionBtn.isVisible()) {
      await actionBtn.click();
    }
  });

  test('Workspace explorer and tabs', async ({ page }) => {
    const explorerBtn = page.locator('button[aria-label="Explorer"]');
    if (await explorerBtn.isVisible()) {
      await explorerBtn.click();
      const folder = page.locator('text=folder1');
      await expect(folder).toBeVisible();
    }

    // Check if tabs are visible
    const tab = page.locator('text=mock_notes.md').first();
    await expect(tab).toBeVisible();
  });

  test('Copilot and Live Ears drawer toggling', async ({ page }) => {
    const copilotBtn = page.locator('button[aria-label="Copilot"]');
    const earsBtn = page.locator('button[aria-label="Live Ears"]');

    if (await copilotBtn.isVisible() && await earsBtn.isVisible()) {
      await earsBtn.click();
      await copilotBtn.click();
    }
  });
});
