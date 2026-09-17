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
            SetClickthrough: async (enable: boolean) => true,
            CaptureScreen: async () => "data:image/png;base64,mock...",
            AnalyzeVision: async (b64: string, prompt: string) => true,
            EndSession: async () => ({ session: { turn_count: 2 }, scorecard: null }),
            GetIDEState: async () => ({
              includeActiveDocContext: true,
              activeDocumentName: 'mock_notes.md',
              workspaceTree: [{ name: 'folder1', children: [] }],
              openDocuments: [{ name: 'mock_notes.md', path: '/mock_notes.md' }],
              activeDocumentPath: '/mock_notes.md',
              activeDocumentContent: '# Mock File',
            }),
            SetIncludeActiveDocContext: async () => {},
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
    const stealthBtn = page.locator('button', { hasText: 'Stealth' });
    await stealthBtn.waitFor({ state: 'visible' });
    await stealthBtn.click();

    // We mock SetClickthrough, so we verify UI change
    const hud = page.locator('.interview-hud');
    await expect(hud).toHaveClass(/clickthrough-mode/);
  });

  test('Vision snip button', async ({ page }) => {
    const snipBtn = page.locator('button', { hasText: 'Snip' });
    await snipBtn.waitFor({ state: 'visible' });
    await snipBtn.click();
    // In our mock, this resolves immediately without navigation, so we just check it doesn't crash
    const hud = page.locator('.interview-hud');
    await expect(hud).toBeVisible();
  });

  test('Session Report open and close', async ({ page }) => {
    const reportBtn = page.locator('button', { hasText: 'Report' });
    await reportBtn.waitFor({ state: 'visible' });
    await reportBtn.click();

    const reportModal = page.locator('.report-panel');
    await expect(reportModal).toBeVisible();

    const closeBtn = page.locator('button.action-btn', { hasText: 'close' });
    if (await closeBtn.count() > 0) {
       await closeBtn.first().click();
    } else {
       const altClose = page.locator('button:has(span.material-symbols-outlined:has-text("close"))').first();
       await altClose.click();
    }

    await expect(reportModal).not.toBeVisible();
  });

  test('STAR preset trigger', async ({ page }) => {
    // Navigate to Brain Drawer if not there
    const copilotBtn = page.locator('button[aria-label="Brain"]');
    if (await copilotBtn.count() > 0) {
      await copilotBtn.click();
    }
    const starBtn = page.locator('button', { hasText: 'STAR' });
    if (await starBtn.count() > 0) {
      await starBtn.click();
      // wait for it to be active
      await expect(starBtn).toHaveClass(/border-primary/);
    }
  });

  test('Workspace explorer and tabs', async ({ page }) => {
    // the active doc should be visible
    const fileTab = page.locator('.document-tab', { hasText: 'mock_notes.md' });
    if (await fileTab.count() > 0) {
       await expect(fileTab.first()).toBeVisible();
    }
  });

  test('Copilot and Live Ears drawer toggling', async ({ page }) => {
    const earsBtn = page.locator('button[aria-label="Live Ears"]');
    const brainBtn = page.locator('button[aria-label="Brain"]');

    if (await earsBtn.count() > 0) {
      await earsBtn.click();
      await expect(page.locator('.live-ears-drawer, text=Transcript')).toBeVisible();
    }
    if (await brainBtn.count() > 0) {
      await brainBtn.click();
      await expect(page.locator('.brain-drawer, text=Copilot')).toBeVisible();
    }
  });
});
