/**
 * ENV VARS required:
 * SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY
 * PADDLE_WEBHOOK_SECRET
 * PADDLE_PRODUCT_INTERVIEW
 * PADDLE_PRODUCT_MENTOR
 * PADDLE_PRODUCT_COUNSEL
 * PADDLE_PRODUCT_CLINIC
 */

import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
};

function getProductMode(productId: string): string | null {
  if (productId === Deno.env.get("PADDLE_PRODUCT_INTERVIEW")) return "interview";
  if (productId === Deno.env.get("PADDLE_PRODUCT_MENTOR")) return "mentor";
  if (productId === Deno.env.get("PADDLE_PRODUCT_COUNSEL")) return "counsel";
  if (productId === Deno.env.get("PADDLE_PRODUCT_CLINIC")) return "clinic";
  return null;
}

function getIncludedSeconds(productMode: string): number {
  switch (productMode) {
    case "mentor":
      return 3600 * 20; // 20 hours/month
    case "counsel":
      return 3600 * 40; // 40 hours/month
    case "clinic":
      return 3600 * 30; // 30 hours/month
    default:
      return 0;
  }
}

async function verifyPaddleSignature(req: Request, rawBody: string): Promise<boolean> {
  const signatureHeader = req.headers.get("paddle-signature");
  if (!signatureHeader) return false;

  const secret = Deno.env.get("PADDLE_WEBHOOK_SECRET");
  if (!secret) {
    console.error("Missing PADDLE_WEBHOOK_SECRET");
    return false;
  }

  // Parse header like ts=1690000000;h1=abc...
  const parts = signatureHeader.split(";");
  let ts = "";
  let h1 = "";
  for (const part of parts) {
    if (part.startsWith("ts=")) ts = part.substring(3);
    if (part.startsWith("h1=")) h1 = part.substring(3);
  }

  if (!ts || !h1) return false;

  const signedPayload = `${ts}:${rawBody}`;
  const key = await crypto.subtle.importKey(
    "raw",
    new TextEncoder().encode(secret),
    { name: "HMAC", hash: "SHA-256" },
    false,
    ["sign", "verify"]
  );

  const signatureBuffer = await crypto.subtle.sign(
    "HMAC",
    key,
    new TextEncoder().encode(signedPayload)
  );

  const hashArray = Array.from(new Uint8Array(signatureBuffer));
  const hashHex = hashArray.map(b => b.toString(16).padStart(2, "0")).join("");

  return hashHex === h1;
}

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  try {
    const rawBody = await req.text();
    const isValid = await verifyPaddleSignature(req, rawBody);

    if (!isValid) {
      return new Response("Invalid signature", { status: 401, headers: corsHeaders });
    }

    const payload = JSON.parse(rawBody);
    const eventType = payload.event_type;
    const data = payload.data;

    if (!eventType || !data) {
      return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
    }

    const supabaseUrl = Deno.env.get("SUPABASE_URL");
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");

    if (!supabaseUrl || !supabaseServiceKey) {
      console.error("Missing Supabase env vars");
      return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
    }

    const supabase = createClient(supabaseUrl, supabaseServiceKey);

    if (eventType === "transaction.completed") {
      const customerId = data.customer_id;
      const customerEmail = data.customer?.email; // wait, structure could be `data.customer_id`, we might need to get email from data or look up? No, prompt says `customer.id` -> paddle_customer_id, `customer.email` -> use to look up. Let's assume data.customer.email or data.customer_email depending on paddle version? Let's check prompt.
      // prompt says:
      // extract from event data:
      // customer.id -> paddle_customer_id
      // customer.email -> use to look up Supabase user by email
      // items[0].price.product_id -> map to product_mode

      const customer = data.customer || {};
      const paddleCustomerId = data.customer_id || customer.id;
      const email = data.customer_email || customer.email;

      const items = data.items || [];
      const productId = items[0]?.price?.product_id;

      if (!paddleCustomerId || !email || !productId) {
         console.error("Missing expected fields in transaction.completed");
         return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }

      const productMode = getProductMode(productId);
      if (!productMode) {
        console.error("Unknown product ID:", productId);
        return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }

      // Look up user_id
      const { data: userIdData, error: rpcError } = await supabase.rpc("get_user_id_by_email", { user_email: email });
      if (rpcError || !userIdData) {
         console.error("User not found or RPC error:", email, rpcError);
         return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }
      const userId = userIdData;

      const licenseKey = crypto.randomUUID();

      const { error: upsertError } = await supabase
        .from("user_entitlements")
        .upsert({
          user_id: userId,
          product_mode: productMode,
          remaining_sessions: 999999,
          license_key: licenseKey,
          paddle_customer_id: paddleCustomerId
        }, { onConflict: "user_id" });

      if (upsertError) {
        console.error("Upsert error in transaction.completed:", upsertError);
      }

    } else if (eventType === "subscription.activated" || eventType === "subscription.updated") {
      const subscriptionId = data.id;
      const customerId = data.customer_id;
      const customerEmail = data.customer?.email || data.customer_email; // paddle v2 usually has customer_id but could be full object?
      // Prompt says: Extract: subscription_id, customer_id, customer_email, current_billing_period.starts_at, current_billing_period.ends_at, items[0].price.product_id

      // I will read both possible locations just in case
      let email = customerEmail;
      if (data.customer && typeof data.customer === 'object') {
          if (!email) email = data.customer.email;
      }

      const startsAt = data.current_billing_period?.starts_at;
      const endsAt = data.current_billing_period?.ends_at;
      const items = data.items || [];
      const productId = items[0]?.price?.product_id;

      if (!subscriptionId || !customerId || !email || !startsAt || !endsAt || !productId) {
         console.error("Missing expected fields in subscription.activated/updated");
         return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }

      const productMode = getProductMode(productId);
      if (!productMode) {
        console.error("Unknown product ID:", productId);
        return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }

      const includedSeconds = getIncludedSeconds(productMode);

      // Look up user_id
      const { data: userIdData, error: rpcError } = await supabase.rpc("get_user_id_by_email", { user_email: email });
      if (rpcError || !userIdData) {
         console.error("User not found or RPC error:", email, rpcError);
         return new Response(JSON.stringify({ received: true }), { headers: corsHeaders });
      }
      const userId = userIdData;

      const { error: upsertError } = await supabase
        .from("user_entitlements")
        .upsert({
          user_id: userId,
          paddle_subscription_id: subscriptionId,
          paddle_customer_id: customerId,
          billing_period_start: startsAt,
          billing_period_end: endsAt,
          included_seconds: includedSeconds,
          product_mode: productMode,
          usage_seconds: 0
        }, { onConflict: "user_id" });

      if (upsertError) {
        console.error("Upsert error in subscription.activated/updated:", upsertError);
      }

    } else if (eventType === "subscription.canceled") {
      const subscriptionId = data.id;
      const email = data.customer?.email || data.customer_email;

      // If we have email, find user_id. If we only have subscriptionId, we could update by that.
      // But let's just update by paddle_subscription_id directly, it's easier and we don't need email.
      if (subscriptionId) {
        const { error: updateError } = await supabase
          .from("user_entitlements")
          .update({
            remaining_sessions: 0,
            demo_expires_at: new Date().toISOString(),
          })
          .eq("paddle_subscription_id", subscriptionId);

        if (updateError) {
          console.error("Update error in subscription.canceled:", updateError);
        }
      } else {
        console.error("Missing subscription ID in canceled event");
      }
    }

    return new Response(JSON.stringify({ received: true }), {
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });

  } catch (error) {
    console.error("Webhook error:", error);
    return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
  }
});
