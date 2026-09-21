import { createClient, type User } from '@supabase/supabase-js';
import { EventsOn, WindowShow } from '../../wailsjs/runtime/runtime';

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
    console.log("🔥 [ENTITLEMENTS] machineId:", machineId, "userId:", authState.user.id);

    // Check if row exists
    const { data: existingData, error: selectError } = await supabase
      .from("user_entitlements")
      .select("*")
      .eq("user_id", authState.user.id)
      .eq("machine_id", machineId)
      .maybeSingle();

    if (selectError) {
      console.error("🔥 [ENTITLEMENTS] SELECT error:", selectError);
    }

    if (!existingData) {
      // No row → insert fresh with 15-min demo timer
      const demoExpiresAt = new Date(Date.now() + 15 * 60 * 1000).toISOString();
      console.log("🔥 [ENTITLEMENTS] No existing row, inserting with demo_expires_at:", demoExpiresAt);
      const { error: insertError } = await supabase.from("user_entitlements").upsert({
        user_id: authState.user.id,
        machine_id: machineId,
        demo_expires_at: demoExpiresAt,
      }, { onConflict: "user_id,machine_id" });

      if (insertError) {
        console.error("🔥 [ENTITLEMENTS] INSERT error:", insertError);
      }
    } else {
      console.log("🔥 [ENTITLEMENTS] Existing row found, demo_expires_at:", existingData.demo_expires_at);
    }

    // Re-read the final row
    const { data, error: finalError } = await supabase
      .from("user_entitlements")
      .select("*")
      .eq("user_id", authState.user.id)
      .eq("machine_id", machineId)
      .maybeSingle();

    if (finalError) {
      console.error("🔥 [ENTITLEMENTS] Final SELECT error:", finalError);
    }

    if (data) {
      authState.demoExpiresAt = data.demo_expires_at;
      authState.paddleStatus = data.paddle_status;
      authState.planType = data.plan_type;
      authState.userEntitlements = data;
      console.log("🔥 [ENTITLEMENTS] State updated. demo_expires_at:", data.demo_expires_at);
    } else {
      console.error("🔥 [ENTITLEMENTS] No data returned after upsert!");
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
    console.log("🔥 [AUTH] on_auth_complete received!", tokenStr ? "Token length: " + tokenStr.length : "No token");
    if (!tokenStr || !supabase) return;

    const [access_token, refresh_token] = tokenStr.split(":");
    if (!access_token || !refresh_token) {
      console.error("🔥 [AUTH] Missing access or refresh token!");
      return;
    }

    console.log("🔥 [AUTH] Setting session...");
    const { data, error } = await supabase.auth.setSession({ access_token, refresh_token });
    
    if (error) {
      console.error("🔥 [AUTH] setSession error:", error);
    }
    
    if (!error && data.user) {
      console.log("🔥 [AUTH] Session set successfully! User ID:", data.user.id);
      authState.user = data.user;
      authState.accessToken = data.session?.access_token || null;
      try {
        await (window as any).go.main.App.SaveToken({ token: tokenStr });
      } catch (e) {
        console.error("🔥 [AUTH] SaveToken failed", e);
      }
      
      console.log("🔥 [AUTH] Syncing user entitlements...");
      await syncUserEntitlements();
      console.log("🔥 [AUTH] Entitlements synced! isGated should update.");
      
      // Bring Wails app to foreground after browser OAuth
      if (typeof WindowShow !== 'undefined') {
        try {
          WindowShow();
        } catch (e) {
           console.log("Failed to show window", e);
        }
      }
    } else {
      console.error("🔥 [AUTH] data.user is null or missing!", data);
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
