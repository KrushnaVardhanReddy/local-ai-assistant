import { test, expect } from "@playwright/test";

test.describe("Cache UI Panel", () => {

  test.beforeEach(async ({ page }) => {
    await page.goto("http://localhost:1420");
    // Open settings panel ("The Brain")
    await page.getByTestId("settings-btn").click();
    await page.waitForSelector("text=Local Q&A Cache");
  });

  test("displays cache stats section", async ({ page }) => {
    const section = page.locator("text=Local Q&A Cache");
    await expect(section).toBeVisible();
  });

  test("Clear Cache button is disabled when cache is empty", async ({ page }) => {
    // When cache has 0 entries, the button should be disabled
    const clearBtn = page.getByRole("button", { name: /clear cache/i });
    await expect(clearBtn).toBeDisabled();
  });

  test("Clear Cache button shows confirmation dialog", async ({ page }) => {
    // Artificially set cached_pairs > 0 by mocking the API
    await page.route("**/api/cache/stats", route =>
      route.fulfill({ json: { cached_pairs: 5, estimated_tokens_saved: 1280 } })
    );
    await page.reload();
    await page.getByTestId("settings-btn").click();
    const clearBtn = page.getByRole("button", { name: /clear cache/i });
    await expect(clearBtn).toBeEnabled();

    // Listen for the confirmation dialog
    page.once("dialog", dialog => {
      expect(dialog.message()).toContain("Are you sure");
      dialog.dismiss(); // Cancel — should NOT clear
    });
    const boundingBox = await clearBtn.boundingBox();
    if (boundingBox) {
      await page.mouse.click(boundingBox.x + boundingBox.width / 2, boundingBox.y + boundingBox.height / 2);
    } else {
      await clearBtn.click({ force: true });
    }
  });

  test("Clear Cache confirmed calls DELETE /api/cache and refreshes stats", async ({ page }) => {
    let deleteCalled = false;
    await page.route("**/api/cache/stats", route =>
      route.fulfill({ json: { cached_pairs: deleteCalled ? 0 : 5, estimated_tokens_saved: 1280 } })
    );
    await page.route("**/api/cache", route => {
      if (route.request().method() === "DELETE") {
        deleteCalled = true;
        route.fulfill({ json: { deleted: 5, message: "Cleared 5 cached Q&A pairs." } });
      }
    });
    await page.reload();
    await page.getByTestId("settings-btn").click();

    page.once("dialog", dialog => dialog.accept());
    const clearBtn = page.getByRole("button", { name: /clear cache/i });
    const boundingBox = await clearBtn.boundingBox();
    if (boundingBox) {
      await page.mouse.click(boundingBox.x + boundingBox.width / 2, boundingBox.y + boundingBox.height / 2);
    } else {
      await clearBtn.click({ force: true });
    }

    await expect(page.locator("text=0 answers cached")).toBeVisible({ timeout: 5000 });
    expect(deleteCalled).toBe(true);
  });

});
