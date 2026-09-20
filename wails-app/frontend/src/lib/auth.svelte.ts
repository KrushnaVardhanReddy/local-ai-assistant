import { createClient, type User } from '@supabase/supabase-js';
import { EventsOn } from '../../wailsjs/runtime/runtime';

const supabaseUrl = import.meta.env.PUBLIC_SUPABASE_URL;
const supabaseAnonKey = import.meta.env.PUBLIC_SUPABASE_ANON_KEY;

export const supabase = supabaseUrl && supabaseAnonKey
  ? createClient(supabaseUrl, supabaseAnonKey)
  : null;

export const authState = $state({
  authMode: (supabaseUrl ? "saas" : "local") as "local" | "saas",
  productMode: (import.meta.env.VITE_PRODUCT || "interview") as string,

  // BarnOwl AI / Lifetime mode
  licenseStatus: "unchecked" as "unchecked" | "active" | "expired" | "not_activated" | "dev_allowed" | "error" | "demo",
  licenseKey: null as string | null,

  // OAuth / SaaS / Demo mode
  user: null as User | null,
  accessToken: null as string | null,
  demoExpiresAt: null as string | null,
  paddleStatus: null as string | null,
  planType: null as string | null,
  userEntitlements: null as any,
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
});



export async function checkDevAllowlist(): Promise<boolean> {
  if (!supabase) return false;
  try {
    const machineId: string = await (window as any).go.main.App.GetMachineId();
    const { data } = await supabase
      .from("dev_allowlist")
      .select("machine_id")
      .eq("machine_id", machineId)
      .single();
    return data !== null;
  } catch { return false; }
}


export async function initLicenseCheck() {
  if (authState.productMode !== "interview") return;

  if (await checkDevAllowlist()) {
    authState.licenseStatus = "dev_allowed";
    return;
  }

  try {
    const status: string = await (window as any).go.main.App.CheckLicense();
    if (status === "active") {
      authState.licenseStatus = "active";
      return;
    }
  } catch {}

  // If no valid license key, check if they have an active OAuth demo session
  if (authState.user && authState.demoExpiresAt) {
    const exp = new Date(authState.demoExpiresAt).getTime();
    if (exp > Date.now()) {
      authState.licenseStatus = "demo";
      return;
    }
  }

  authState.licenseStatus = "not_activated";
}

export async function activateLicense(key: string) {
  try {
    const success: boolean = await (window as any).go.main.App.ActivateLicense(key);
    if (success) {
      authState.licenseStatus = "active";
      authState.licenseKey = key;
    } else {
      throw new Error("Invalid license key");
    }
  } catch (err: any) {
    throw err;
  }
}


export async function syncUserEntitlements() {
  if (!supabase || !authState.user) return;
  try {
    const machineId: string = await (window as any).go.main.App.GetMachineId();

    // First, check if row exists
    const { data: existingData } = await supabase
      .from("user_entitlements")
      .select("*")
      .eq("user_id", authState.user.id)
      .eq("machine_id", machineId)
      .single();

    if (!existingData) {
      // Upsert new row with demo_expires_at 15 mins from now
      const demoExpiresAt = new Date(Date.now() + 15 * 60 * 1000).toISOString();
      await supabase.from("user_entitlements").upsert({
        user_id: authState.user.id,
        machine_id: machineId,
        demo_expires_at: demoExpiresAt,
      });
    } else {
      // Just an upsert in case we need to update updated_at or similar in the future,
      // but otherwise existing is fine
      await supabase.from("user_entitlements").upsert({
        user_id: authState.user.id,
        machine_id: machineId,
      });
    }

    const { data } = await supabase
      .from("user_entitlements")
      .select("*")
      .eq("user_id", authState.user.id)
      .eq("machine_id", machineId)
      .single();

    if (data) {
      authState.demoExpiresAt = data.demo_expires_at;
      authState.paddleStatus = data.paddle_status;
      authState.planType = data.plan_type;
      authState.userEntitlements = data;
    }

    if (authState.productMode === "interview") {
      await initLicenseCheck();
    }
  } catch (err) {
    console.error("Failed to sync user entitlements", err);
  }
}

export function initAuthEventListeners() {
  const onEvent = typeof window !== "undefined" && (window as any).runtime?.EventsOn ? (window as any).runtime.EventsOn : EventsOn;
  onEvent("on_auth_complete", async (tokenStr: string) => {
    if (!tokenStr || !supabase) return;

    const [access_token, refresh_token] = tokenStr.split(":");
    if (!access_token || !refresh_token) return;

    const { data, error } = await supabase.auth.setSession({ access_token, refresh_token });
    if (!error && data.user) {
      authState.user = data.user;
      authState.accessToken = data.session?.access_token || null;
      try {
        await (window as any).go.main.App.SaveToken({ token: tokenStr });
      } catch (e) {}
      await syncUserEntitlements();
    }
  });
}

if (typeof window !== "undefined") {
  initAuthEventListeners();
  (window as any).__authState = authState;
}

export const cloudAuthState = $state({
  apiKey: typeof window !== 'undefined' ? localStorage.getItem('cloud_api_key') : null
});

export function setCloudApiKey(key: string | null) {
  cloudAuthState.apiKey = key;
  if (typeof window !== 'undefined') {
    if (key === null) {
      localStorage.removeItem('cloud_api_key');
    } else {
      localStorage.setItem('cloud_api_key', key);
    }
  }
}

export async function restoreSession() {
  if (authState.authMode === "local" || !supabase) return;

  try {
    const tokenStr: string = await (window as any).go.main.App.LoadToken();
    if (!tokenStr) return;

    const [access_token, refresh_token] = tokenStr.split(":");
    if (!access_token || !refresh_token) return;

    const { data, error } = await supabase.auth.setSession({ access_token, refresh_token });

    if (error || !data.user) {
      console.warn("Session restore failed, clearing token", error);
      await (window as any).go.main.App.DeleteToken();
      authState.user = null;
      authState.accessToken = null;
    } else {
      authState.user = data.user;
      authState.accessToken = data.session?.access_token || null;
    }
  } catch (err) {
    // Normal for first launch (no token found)
    console.debug("No session found in keychain");
  }
}

export async function signUp(email: string, password: string) {
  if (authState.authMode === "local" || !supabase) return { error: "Local mode" };

  const { data, error } = await supabase.auth.signUp({ email, password });
  if (error) {
    return { error: error.message };
  }
  return { error: null, data };
}

export async function signIn(email: string, password: string) {
  if (authState.authMode === "local" || !supabase) return { error: "Local mode" };

  const { data, error } = await supabase.auth.signInWithPassword({ email, password });
  if (error) {
    return { error: error.message };
  }

  if (data.session) {
    const token = `${data.session.access_token}:${data.session.refresh_token}`;
    try {
      await (window as any).go.main.App.SaveToken({ token });
      authState.user = data.user;
      authState.accessToken = data.session.access_token;
    } catch (err) {
      console.error("Failed to save token to keychain", err);
      return { error: "Failed to save session securely" };
    }
  }

  return { error: null };
}

export async function signOut() {
  if (authState.authMode === "local" || !supabase) return;

  await supabase.auth.signOut();

  try {
    await (window as any).go.main.App.DeleteToken();
  } catch (err) {
    console.error("Failed to delete token from keychain", err);
  }

  authState.user = null;
  authState.accessToken = null;
}

if (supabase) {
  supabase.auth.onAuthStateChange(async (event, session) => {
    if (event === 'TOKEN_REFRESHED' && session) {
      authState.user = session.user;
      authState.accessToken = session.access_token;
      const token = `${session.access_token}:${session.refresh_token}`;
      try {
        await (window as any).go.main.App.SaveToken({ token });
      } catch (err) {
        console.error("Failed to update token in keychain", err);
      }
    } else if (event === 'SIGNED_OUT') {
      authState.user = null;
      authState.accessToken = null;
      try {
        await (window as any).go.main.App.DeleteToken();
      } catch (err) {
        console.error("Failed to delete token from keychain", err);
      }
    }
  });
}
