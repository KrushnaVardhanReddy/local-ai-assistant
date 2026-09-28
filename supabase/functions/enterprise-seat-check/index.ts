// SECURITY NOTE:
// BYPASS_ENTERPRISE_CHECKS=true is ONLY set in the local dev Supabase project.
// The production Supabase project MUST NOT have this variable set.
// If somehow both SUPABASE_URL (prod) and BYPASS_ENTERPRISE_CHECKS=true are
// present, this function will REFUSE to bypass and will return a 403.
// This prevents accidental bypass in production.

import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const PROD_URL_FRAGMENT = ".supabase.co"; // prod URLs always contain this

Deno.serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: { "Access-Control-Allow-Origin": "*" } });
  }

  const supabaseUrl = Deno.env.get("SUPABASE_URL") ?? "";
  const bypassEnabled = Deno.env.get("BYPASS_ENTERPRISE_CHECKS") === "true";
  const isProduction = supabaseUrl.includes(PROD_URL_FRAGMENT);

  // Hard block: bypass cannot be active in production
  if (bypassEnabled && isProduction) {
    console.error("[SECURITY] BYPASS_ENTERPRISE_CHECKS is set in PRODUCTION. Refusing.");
    return new Response(JSON.stringify({ error: "Configuration error" }), {
      status: 403,
      headers: { "Content-Type": "application/json" },
    });
  }

  const { email } = await req.json();
  if (!email || !email.includes("@")) {
    return new Response(JSON.stringify({ error: "Invalid email" }), { status: 400 });
  }

  const domain = email.split("@")[1].toLowerCase();

  // DEV bypass: return a fake org for testing without a real DB record
  if (bypassEnabled && !isProduction) {
    console.warn(`[DEV] Bypassing enterprise check for domain: ${domain}`);
    return new Response(JSON.stringify({
      found: true,
      org: { id: "dev-bypass-org", name: "Dev Bypass Org", max_seats: 999, seats_used: 0, status: "active" },
    }), { status: 200, headers: { "Content-Type": "application/json" } });
  }

  const supabase = createClient(
    supabaseUrl,
    Deno.env.get("SUPABASE_SERVICE_ROLE_KEY") ?? "",
  );

  // Look up org by email domain
  const { data: org, error } = await supabase
    .from("organizations")
    .select("id, name, max_seats, status")
    .eq("email_domain", domain)
    .eq("status", "active")
    .single();

  if (error || !org) {
    return new Response(JSON.stringify({ found: false, error: "No active enterprise license found for this domain." }), {
      status: 404,
      headers: { "Content-Type": "application/json" },
    });
  }

  // Check seat availability
  const { count } = await supabase
    .from("user_entitlements")
    .select("*", { count: "exact", head: true })
    .eq("org_id", org.id)
    .eq("is_enterprise", true);

  const seatsUsed = count ?? 0;
  if (seatsUsed >= org.max_seats) {
    return new Response(JSON.stringify({ found: true, error: `All ${org.max_seats} seats are in use. Contact your admin.` }), {
      status: 403,
      headers: { "Content-Type": "application/json" },
    });
  }

  return new Response(JSON.stringify({
    found: true,
    org: { id: org.id, name: org.name, max_seats: org.max_seats, seats_used: seatsUsed, status: org.status },
  }), { status: 200, headers: { "Content-Type": "application/json" } });
});
