import sys

content = open("wails-app/frontend/src/lib/auth.svelte.ts").read()

content = content.replace("""export const authState = $state({
  authMode: (supabaseUrl ? "saas" : "local") as "local" | "saas",
  productMode: (import.meta.env.VITE_PRODUCT || "interview") as string,

  // BarnOwl AI / Lifetime mode
  licenseStatus: "unchecked" as "unchecked" | "active" | "expired" | "not_activated" | "dev_allowed" | "error" | "demo",
  licenseKey: null as string | null,

  // OAuth / SaaS / Demo mode
  user: null as User | null,
  accessToken: null as string | null,
  demoExpiresAt: null as string | null,
  stripeStatus: null as string | null,
  planType: null as string | null,
});""", """export const authState = $state({
  authMode: (supabaseUrl ? "saas" : "local") as "local" | "saas",
  productMode: (import.meta.env.VITE_PRODUCT || "interview") as string,

  // BarnOwl AI / Lifetime mode
  licenseStatus: "unchecked" as "unchecked" | "active" | "expired" | "not_activated" | "dev_allowed" | "error" | "demo",
  licenseKey: null as string | null,

  // OAuth / SaaS / Demo mode
  user: null as User | null,
  accessToken: null as string | null,
  demoExpiresAt: null as string | null,
  stripeStatus: null as string | null,
  planType: null as string | null,
});

$effect.root(() => {
  $effect(() => {
    if (typeof window !== "undefined" && (window as any).go?.main?.App) {
      if (authState.licenseStatus === "demo" && authState.accessToken) {
        (window as any).go.main.App.SetProxyToken(authState.accessToken);
      } else {
        (window as any).go.main.App.SetProxyToken("");
      }
    }
  });
});""")

open("wails-app/frontend/src/lib/auth.svelte.ts", "w").write(content)
