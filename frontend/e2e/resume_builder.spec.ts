import { test, expect } from '@playwright/test';

test.describe('Resume Builder Component', () => {
  test('should render editor, preview pane and themes', async ({ page, isMobile }) => {
    test.skip(isMobile, 'Skip desktop-only resume builder on mobile');
    await page.goto('/');

    // Ensure the toggle button exists and click it
    const toggleBtn = page.getByRole('button', { name: 'Toggle Resume Builder' });
    await expect(toggleBtn).toBeVisible();
    await toggleBtn.click();

    // Verify the resume builder container appears
    const resumeContainer = page.locator('.resume-builder-container');
    await expect(resumeContainer).toBeVisible();

    // Verify the "Tailor to Job" button exists
    const tailorBtn = page.getByRole('button', { name: /TAILOR TO JOB/i });
    await expect(tailorBtn).toBeVisible();

    // Verify the theme selector exists
    const themeSelect = page.locator('select');
    await expect(themeSelect).toBeVisible();

    // Test the text area and markdown render
    const editor = page.locator('textarea');
    await expect(editor).toBeVisible();

    await editor.fill('# My Resume');
    const previewPane = page.locator('.markdown-preview');
    await expect(previewPane).toContainText('My Resume');

    // Test PDF export button
    let printCalled = false;
    await page.exposeFunction('onPrint', () => {
      printCalled = true;
    });
    await page.evaluate(() => {
      window.print = window.onPrint;
    });

    const pdfBtn = page.getByRole('button', { name: /PDF/i });
    // PDF button becomes enabled when editor has text
    await expect(pdfBtn).toBeEnabled();

    await pdfBtn.click();

    // Since tick and window.print are asynchronous in svelte, we just wait a bit or evaluate
    // to check the mock was called.
    await expect(async () => {
      expect(printCalled).toBe(true);
    }).toPass();
  });
});
