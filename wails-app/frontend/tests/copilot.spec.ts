import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

test.describe.serial('Copilot UI Tests', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).go = {
        backend: {
          App: {
            SendChat: async (msg: string) => {
              if ((window as any).mockSendChat) {
                (window as any).mockSendChat(msg);
              }
            }
          }
        },
        main: {
          App: {
            GetIDEState: async () => ({ includeActiveDocContext: false }),
            GetSystemStatus: async () => ({ llm_provider: 'auto' }),
            GetAudioDevices: async () => ([]),
            CheckLicense: async () => ('error'),
            LoadToken: async () => (''),
            GetMachineId: async () => ('test-machine-id'),
            SetIncludeActiveDocContext: async () => (true),
            ClearState: async () => (true),
            SendCustomPrompt: async () => (true),
            SetClickthrough: async (val: boolean) => { return true; },
            SendChat: async (msg: string) => {
              if ((window as any).mockSendChat) {
                (window as any).mockSendChat(msg);
              }
            }
          }
        }
      };

      (window as any).go.backend = (window as any).go.main;

      (window as any).runtime = {
        EventsOn: () => {},
        WindowSetSize: () => {},
        WindowCenter: () => {},
        EventsEmit: () => {}
      };

      if (!(window as any).__authState) {
        (window as any).__authState = { authMode: 'local' };
      }
    });

    await page.route('**/chat', async (route) => {
      const req = route.request();
      if (req.method() === 'POST') {
        const body = req.postDataJSON();
        if (body && body.message) {
           await page.evaluate((msg) => {
               if ((window as any).mockSendChat) {
                   (window as any).mockSendChat(msg);
               }
           }, body.message);
        }
      }
      await route.fulfill({ status: 200, json: {} });
    });

    await page.goto('/');

    await page.waitForSelector('.app-shell');
    await page.waitForFunction(() => (window as any).__authState !== undefined);

    await page.evaluate(() => {
      if ((window as any).__authState) {
        (window as any).__authState.authMode = 'local';
      }
    });
  });

  test('STAR Method preset button calls SendChat with correct prompt', async ({ page }) => {
    let sentMessage = '';
    await page.exposeFunction('mockSendChat', (msg: string) => {
        sentMessage = msg;
    });

    // Make sure CopilotDrawer is open because `.star-primed` rendering is guaranteed there
    const copilotBtn = page.locator('button[data-testid="activity-bar-copilot"]');
    if (await copilotBtn.count() > 0) {
        await copilotBtn.click();
    }

    const starBtn = page.locator('button', { hasText: 'STAR' }).first();
    await starBtn.waitFor({ state: 'visible', timeout: 5000 });

    await starBtn.click();

    // Also click the ConvPanel button if present, because InterviewHUD's button uses `App.SendChat`
    // which guarantees `mockSendChat` fires reliably without WebSocket mocking constraints.
    const convBtn = page.locator('.conv-panel-wrapper button', { hasText: 'STAR' }).first();
    if (await convBtn.isVisible()) {
        await convBtn.click();
    }

    await expect.poll(() => sentMessage, { timeout: 5000 }).toBe('Format your next response using the STAR method (Situation, Task, Action, Result).');

    // Playwright css + text pseudo-selector must be separated
    const starPrimedClass = page.locator('.star-primed').first();
    const starMethodText = page.locator('text="STAR Method Primed"').first();

    const isVisible = await Promise.race([
        starPrimedClass.waitFor({ state: 'visible', timeout: 5000 }).then(() => true).catch(() => false),
        starMethodText.waitFor({ state: 'visible', timeout: 5000 }).then(() => true).catch(() => false)
    ]);

    expect(isVisible).toBe(true);
  });

  test('Stealth mode toggles clickthrough classes', async ({ page }) => {
    const stealthBtn = page.locator('button[title="Stealth Mode (Clickthrough)"]');
    await stealthBtn.waitFor({ state: 'visible', timeout: 5000 });

    const hud = page.locator('.interview-hud');

    await stealthBtn.click();
    await expect(hud).toHaveClass(/clickthrough-mode/);

    await stealthBtn.click();
    await expect(hud).not.toHaveClass(/clickthrough-mode/);
  });
});
