import { test, expect } from '@playwright/test';

test.describe.serial('Mock Interview Mode Tests', () => {
  test.beforeEach(async ({ page }) => {
    // Inject mock wails/go objects before scripts run to avoid Vite waiting for Wails forever
    await page.addInitScript(() => {
      (window as any).go = {
        main: {
          App: {
            GetIDEState: async () => ({
              includeActiveDocContext: false,
              isMockMode: false,
              transcriptHistory: [],
              pendingTranscripts: [],
              response: {
                chunks: [],
                chips: []
              }
            }),
            GetSystemStatus: async () => ({ llm_provider: 'auto' }),
            GetAudioDevices: async () => ([]),
            CheckLicense: async () => ('active'),
            LoadToken: async () => (''),
            GetMachineId: async () => ('test-machine-id'),
            ToggleMockMode: async () => {
              const state = await (window as any).go.main.App.GetIDEState();
              state.isMockMode = !state.isMockMode;
              return state.isMockMode;
            },
            SendChat: async (msg) => {
              // Mock backend receiving chat
              return true;
            }
          }
        }
      };
      (window as any).runtime = {
        EventsOn: (eventName, callback) => {
          if (!(window as any)._events) (window as any)._events = {};
          if (!(window as any)._events[eventName]) (window as any)._events[eventName] = [];
          (window as any)._events[eventName].push(callback);
        },
        EventsEmit: (eventName, ...args) => {
          if ((window as any)._events && (window as any)._events[eventName]) {
            (window as any)._events[eventName].forEach(cb => cb(...args));
          }
        },
        WindowSetSize: () => {},
        WindowCenter: () => {}
      };

      if (!(window as any).__authState) {
        (window as any).__authState = {
          authMode: 'saas',
          productMode: 'interview',
          licenseStatus: 'active'
        };
      }
    });

    await page.goto('/');
    await page.waitForSelector('.interview-hud');
  });

  test('Toggle Mock Interview mode and assert UI state', async ({ page }) => {
    // Ensure we start with the mock mode button visible
    const mockBtn = page.locator('button[title="Mock Mode"]');
    await expect(mockBtn).toBeVisible();

    // The button shouldn't have the active class initially
    await expect(mockBtn).not.toHaveClass(/active/);

    // Expose a function to change the IDEState mock
    await page.evaluate(() => {
      let isMockMode = false;
      const originalGetIDEState = (window as any).go.main.App.GetIDEState;

      // Update our stored state on ToggleMockMode
      (window as any).go.main.App.ToggleMockMode = async () => {
        isMockMode = !isMockMode;
        return isMockMode;
      };

      (window as any).go.main.App.GetIDEState = async () => {
        const state = await originalGetIDEState();
        state.isMockMode = isMockMode;
        return state;
      };
    });

    // Click to toggle mock mode
    await mockBtn.click();

    // Check if the mock button has the active class
    await expect(mockBtn).toHaveClass(/active/);

    // Instead of messing with Reactivity timings, emit the exact Wails event the websocket module is bound to:
    // It's called `on_transcript` based on ws.svelte

    await page.evaluate(() => {
      const newTranscript = {
          text: "Can you explain your experience with React?",
          role: "interviewer"
      };
      (window as any).runtime.EventsEmit("on_transcript", newTranscript);
    });

    const interviewerBubble = page.locator('.bubble-interviewer').first();
    await expect(interviewerBubble).toBeVisible({ timeout: 5000 });

    const transcriptText = page.locator('text="Can you explain your experience with React?"').first();
    await expect(transcriptText).toBeVisible();
  });
});
