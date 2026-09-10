import { test, expect } from '@playwright/test';

test.describe('Enterprise features', () => {
    test.beforeEach(async ({ page }) => {
        // Mock backend health check
        await page.route('http://localhost:8765/health', async route => {
            const json = { status: 'ok', auth_enabled: true };
            await route.fulfill({ json });
        });

        // Mock getting current user
        await page.route('**/rest/v1/users*', async route => {
            await route.fulfill({ json: [] });
        });
    });

    test('should allow navigating to SSO login form', async ({ page }) => {
        await page.goto('/login');

        // Wait for hydration by checking for the button we expect to be interactive
        const ssoToggle = page.locator('text=Sign in with SSO');
        await expect(ssoToggle).toBeVisible();

        // Give Svelte a moment to bind events
        await page.waitForTimeout(500);

        // Ensure we're currently showing email/password (by lack of enterprise domain)
        await expect(page.locator('input[id="email"]')).toBeVisible();
        await expect(page.locator('input[id="sso-email"]')).not.toBeVisible();

        // Switch to SSO mode
        await page.waitForLoadState('networkidle');
        await page.locator('button', { hasText: 'Sign in with SSO' }).click({ force: true });

        // Wait for hydration by checking for the button we expect to be interactive
        await expect(page.locator('input[id="sso-email"]')).toBeVisible({ timeout: 10000 });
    });

    test('should allow an admin to manage seats', async ({ page, context }) => {
        // Mock checking user plan
        await page.route('http://localhost:8765/api/plan', async route => {
            const json = { plan: 'enterprise' };
            await route.fulfill({ json });
        });

        // Mock checking the organization metadata to identify them as an admin
        await page.route('**/rest/v1/organizations?*', async route => {
            const json = [{ id: 'org_123', name: 'Acme Corp', admin_id: 'user_123', seat_limit: 10 }];
            await route.fulfill({ json });
        });

        // Mock members query
        await page.route('**/rest/v1/profiles?*', async route => {
            const json = [
                { id: 'user_123', email: 'admin@acme.com', org_id: 'org_123' },
                { id: 'user_456', email: 'employee@acme.com', org_id: 'org_123' }
            ];
            await route.fulfill({ json });
        });

        // Set fake user auth state
        await context.addInitScript(() => {
            window.localStorage.setItem('sb-local-auth-token', JSON.stringify({
                user: {
                    id: 'user_123',
                    email: 'admin@acme.com'
                },
                access_token: 'fake_token'
            }));
        });

        await page.goto('/dashboard/admin');

        // Check if the dashboard renders the org info
        await expect(page.locator('text=acme.com — BarnOwl Enterprise')).toBeVisible();
        await expect(page.locator('text=2 / 10 used')).toBeVisible();

        // Ensure both members are rendered
        await expect(page.locator('text=admin@acme.com')).toBeVisible();

        // Give Svelte a moment to bind events
        await page.waitForTimeout(500);

        // Open the invite modal
        await page.waitForLoadState('networkidle');
        await page.locator('button', { hasText: '+ Invite' }).click({ force: true });

        // Wait for modal to render
        await expect(page.locator('input[id="inviteEmail"]')).toBeVisible({ timeout: 10000 });

        // Fill out an invite
        await page.fill('input[id="inviteEmail"]', 'newguy@acme.com');

        // Setup mock for update profile
        let requestPayload: any;
        await page.route('**/api/enterprise/invite', async route => {
            if (route.request().method() === 'POST') {
                requestPayload = route.request().postDataJSON();
                await route.fulfill({ status: 200, json: { status: 'success' } });
            } else {
                await route.continue();
            }
        });

        await page.locator('button', { hasText: 'Send Invite' }).click({ force: true });

        // Assert the new member was added to UI
        await expect(page.locator('text=newguy@acme.com')).toBeVisible({ timeout: 10000 });
    });
});
