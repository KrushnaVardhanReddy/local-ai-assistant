1. **Update `authState` in `wails-app/frontend/src/lib/auth.svelte.ts`**
   - Add entitlement fields to the `$state` object: `byok_pass_active: false`, `remaining_sessions: 0`, `plan_type: "none"`.
   - Add an exported async function `fetchEntitlement()`:
     ```ts
     const table = import.meta.env.VITE_SUPABASE_PROFILES_TABLE;
     if (!table) return;
     const { data, error } = await supabase.from(table).select('pass_expires_at, remaining_sessions, plan_type').eq('id', authState.user.id).single();
     ```
   - Calculate `byok_pass_active = data && data.pass_expires_at ? new Date(data.pass_expires_at) > new Date() : false`.
   - In `auth.svelte.ts`, modify `restoreSession`, `signIn`, and the `onAuthStateChange` hook (verified at line 107) to call `fetchEntitlement()` after successful auth.

2. **Verify Changes to `auth.svelte.ts`**
   - Run `cat wails-app/frontend/src/lib/auth.svelte.ts` to ensure the modifications were applied correctly.

3. **Update `wails-app/frontend/src/lib/Assistant.svelte`**
   - In `Assistant.svelte`, add a derived `isGated` boolean:
     ```ts
     const isGated = $derived(authState.authMode === 'saas' && (!authState.user || (!authState.byok_pass_active && authState.remaining_sessions <= 0)));
     ```
   - In `Assistant.svelte`, add a prominent overlay when `isGated` is true. The overlay displays the "Pass Expired / License Required" lock banner and a button that opens the `AuthModal` (if logged out) or redirects to `import.meta.env.VITE_BILLING_URL || ''`.
   - In `Assistant.svelte`, add a check in its `handleKeydown` to return early if `isGated` is true. Also disable the `<textarea>` (verified at line 727) by adding `disabled={isGated}` to it.

4. **Verify Changes to `Assistant.svelte`**
   - Run `cat wails-app/frontend/src/lib/Assistant.svelte` to ensure gating logic and UI elements are correctly placed.

5. **Update `wails-app/frontend/src/lib/Settings.svelte`**
   - Replace any references to `authState.plan` with `authState.plan_type`. For example, `disabled={authState.plan_type !== 'lifetime'}`.
   - In the `Settings.svelte` file, locate the `<h2>Account</h2>` section (verified at line 533). Inside this section, below the email or badges (verified at line 540-542), display the current active pass status. If `authState.byok_pass_active` is true, show "Active: BYOK Pass". If `authState.remaining_sessions > 0`, show `SaaS Managed — ${authState.remaining_sessions} sessions remaining`.
   - Add a `<button>` element with the label "Recharge Pass" linking to `window.open(import.meta.env.VITE_BILLING_URL || '', '_blank')` in the `<div class="actions">` (verified at line 544).

6. **Verify Changes to `Settings.svelte`**
   - Run `cat wails-app/frontend/src/lib/Settings.svelte` to confirm the Account section changes and references to `plan_type`.

7. **Add Playwright Tests (`wails-app/frontend/tests/entitlement.spec.ts`)**
   - Create `wails-app/frontend/tests/entitlement.spec.ts`.
   - Add test cases to mount the app and mock the Supabase network API requests using `page.route` to simulate different user entitlement profiles (unlicensed vs. licensed).
   - Test that the gating banner appears when entitlement is missing, and the input/hotkeys are disabled.
   - Run `cat wails-app/frontend/tests/entitlement.spec.ts` to verify the file was created correctly.

8. **Execute Frontend Tests**
   - Run `cd wails-app/frontend && npx playwright test` to execute the Playwright tests and ensure no regressions.

9. **Complete Pre-Commit Steps**
   - Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.

10. **Submit**
   - Call the `submit` tool to finish the task.
