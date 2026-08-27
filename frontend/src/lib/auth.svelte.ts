import { invoke } from '@tauri-apps/api/core';
import { createClient, type User } from '@supabase/supabase-js';

const supabaseUrl = import.meta.env.PUBLIC_SUPABASE_URL;
const supabaseAnonKey = import.meta.env.PUBLIC_SUPABASE_ANON_KEY;

export const supabase = supabaseUrl && supabaseAnonKey
  ? createClient(supabaseUrl, supabaseAnonKey)
  : null;

export const authState = $state({
  user: null as User | null,
  accessToken: null as string | null,
  authMode: (supabaseUrl ? "saas" : "local") as "local" | "saas"
});

export async function restoreSession() {
  if (authState.authMode === "local" || !supabase) return;

  try {
    const tokenStr: string = await invoke("load_token");
    if (!tokenStr) return;

    const [access_token, refresh_token] = tokenStr.split(":");
    if (!access_token || !refresh_token) return;

    const { data, error } = await supabase.auth.setSession({ access_token, refresh_token });

    if (error || !data.user) {
      console.warn("Session restore failed, clearing token", error);
      await invoke("delete_token");
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

export async function signIn(email: string, password: string) {
  if (authState.authMode === "local" || !supabase) return { error: "Local mode" };

  const { data, error } = await supabase.auth.signInWithPassword({ email, password });
  if (error) {
    return { error: error.message };
  }

  if (data.session) {
    const token = `${data.session.access_token}:${data.session.refresh_token}`;
    try {
      await invoke("save_token", { token });
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
    await invoke("delete_token");
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
        await invoke("save_token", { token });
      } catch (err) {
        console.error("Failed to update token in keychain", err);
      }
    } else if (event === 'SIGNED_OUT') {
      authState.user = null;
      authState.accessToken = null;
      try {
        await invoke("delete_token");
      } catch (err) {
        console.error("Failed to delete token from keychain", err);
      }
    }
  });
}
