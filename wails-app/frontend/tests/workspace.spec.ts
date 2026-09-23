import { test, expect } from '@playwright/test';

test.describe.serial('Core IDE & File System', () => {
  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      (window as any).__authState = { authMode: 'local', productMode: 'interview' };

      const mockWorkspaceTree = [
        { name: 'mock_dir', path: '/mock_dir', isDirectory: true, isDir: true, children: [
            { name: 'test_file.txt', path: '/mock_dir/test_file.txt', isDirectory: false, isDir: false }
        ]}
      ];

      const mockIDEState = {
        includeActiveDocContext: false,
        workspaceTree: mockWorkspaceTree,
        openDocuments: [],
        activeDocumentPath: '',
        activeDocumentContent: '',
        script: ''
      };

      // Mock both window.go.backend and window.go.main for compatibility with the test request and current app architecture
      const appBindings = {
        GetIDEState: async () => mockIDEState,
        GetSystemStatus: async () => ({ llm_provider: 'auto' }),
        GetAudioDevices: async () => ([]),
        CheckLicense: async () => ('active'),
        LoadToken: async () => (''),
        GetMachineId: async () => ('test-machine-id'),
        GetIndexedPaths: async () => ([]),
        PromptOpenFile: async () => ({ path: '/mock_dir/test_file.txt', content: '# Mock Content\nThis is a mock file.' }),
        HideFromTaskbar: async () => {},
        ToggleMic: async () => true,
        OpenFile: async (path: string) => {
           if (typeof (window as any).mockOpenFile === 'function') {
              (window as any).mockOpenFile(path);
           }
           return true;
        }
      };

      (window as any).go = {
        backend: { App: appBindings },
        main: { App: appBindings }
      };

      (window as any).runtime = {
        EventsOn: () => {},
        WindowSetSize: () => {},
        WindowCenter: () => {},
        WindowSetAlwaysOnTop: () => {},
        WindowShow: () => {}
      };
    });

    await page.goto('/');

    // Bypass auth modal completely
    await page.evaluate(() => {
      const auth = (window as any).__authState;
      if (auth) {
        auth.authMode = 'local';
        auth.productMode = 'interview';
        auth.licenseStatus = 'active';
      }
      const modal = document.querySelector('.fixed.inset-0.z-\\[9999\\]');
      if (modal) modal.remove();
    });

    await page.waitForSelector('.app-shell', { state: 'visible' });
  });

  test('Clicking a file in FileTree opens a new tab and renders content', async ({ page }) => {
    let openFileCalled = false;
    let openedPath = '';
    await page.exposeFunction('mockOpenFile', (path: string) => {
        openFileCalled = true;
        openedPath = path;
    });

    // Re-mock state after open to trigger Svelte reactions
    await page.evaluate(() => {
        const originalOpenFile = window.go.main.App.OpenFile;
        const mockOpenFn = async (path: string) => {
            await originalOpenFile(path);

            const newState = {
              includeActiveDocContext: false,
              workspaceTree: [
                { name: 'mock_dir', path: '/mock_dir', is_dir: true, children: [
                    { name: 'test_file.txt', path: '/mock_dir/test_file.txt', is_dir: false }
                ]}
              ],
              openDocuments: [{ name: 'test_file.txt', path: '/mock_dir/test_file.txt' }],
              activeDocumentPath: '/mock_dir/test_file.txt',
              activeDocumentContent: '# Mock Content\nThis is a mock file.'
            };

            window.go.main.App.GetIDEState = async () => newState;
            window.go.backend.App.GetIDEState = async () => newState;

            return true;
        };

        window.go.main.App.OpenFile = mockOpenFn;
        window.go.backend.App.OpenFile = mockOpenFn;
    });

    const explorerBtn = page.locator('button[data-testid="activity-bar-explorer"]');
    await explorerBtn.waitFor({ state: 'visible' });
    await explorerBtn.click();

    const fileTree = page.locator('.workspace-sidebar-container');
    await expect(fileTree).toBeVisible();

    const dirNode = page.locator('.tree-node .name', { hasText: 'mock_dir' }).first();
    await expect(dirNode).toBeVisible({ timeout: 5000 });
    await dirNode.click();

    const fileNode = page.locator('.tree-node .name', { hasText: 'test_file.txt' }).first();
    await expect(fileNode).toBeVisible({ timeout: 5000 });
    await fileNode.click();

    await expect.poll(() => openFileCalled, { timeout: 5000 }).toBeTruthy();
    expect(openedPath).toBe('/mock_dir/test_file.txt');

    const newTab = page.locator('.tab .name', { hasText: 'test_file.txt' }).first();
    await expect(newTab).toBeVisible({ timeout: 5000 });

    const editorContent = page.locator('.cm-content', { hasText: 'Mock Content' }).first();
    await expect(editorContent).toBeVisible({ timeout: 5000 });
  });
});
