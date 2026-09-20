import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2.39.3";

// Initialize Supabase Admin client
const supabaseUrl = Deno.env.get("SUPABASE_URL");
const supabaseServiceRoleKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");

const supabase = supabaseUrl && supabaseServiceRoleKey
  ? createClient(supabaseUrl, supabaseServiceRoleKey)
  : null;

// Function to verify HMAC-SHA256 signature
async function verifySignature(signatureHeader: string | null, rawBody: string, secret: string | undefined): Promise<boolean> {
  if (!signatureHeader || !secret) return false;

  const parts = signatureHeader.split(";");
  let ts = "";
  let h1 = "";

  for (const part of parts) {
    if (part.startsWith("ts=")) ts = part.substring(3);
    else if (part.startsWith("h1=")) h1 = part.substring(3);
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

serve(async (req: Request) => {
  if (!supabase) {
    console.error("Supabase environment variables missing");
    return new Response("Internal Server Error", { status: 500 });
  }

  const rawBody = await req.text();
  const signatureHeader = req.headers.get("paddle-signature");
  const secret = Deno.env.get("PADDLE_WEBHOOK_SECRET");

  const isValid = await verifySignature(signatureHeader, rawBody, secret);
  if (!isValid) {
    return new Response("Unauthorized", { status: 401 });
  }

  let body;
  try {
    body = JSON.parse(rawBody);
  } catch (e) {
    return new Response("Invalid JSON", { status: 400 });
  }

  const eventType = body.event_type;
  const data = body.data;

  if (!eventType || !data) {
    return new Response("OK", { status: 200 }); // Paddle requires 200 OK
  }

  try {
    // We expect the custom_data to contain the user_id for matching the entitlements
    const customerId = data.customer_id;
    let userId = data.custom_data?.user_id;

    if (!userId && data.customer?.custom_data?.user_id) {
        userId = data.customer.custom_data.user_id;
    }

    if (!userId) {
        // We will try fetching user_id by customer_id if it's a recurrent payment
        if (customerId) {
            const { data: entitlement } = await supabase
                .from("user_entitlements")
                .select("user_id")
                .eq("paddle_customer_id", customerId)
                .maybeSingle();

            if (entitlement && entitlement.user_id) {
                userId = entitlement.user_id;
            }
        }
    }

    if (!userId) {
        console.error("No user_id found in event custom_data or existing entitlements for customer:", customerId);
        return new Response("OK", { status: 200 }); // Acknowledge to stop retries
    }

    switch (eventType) {
      case "transaction.completed": {
        const items = data.items || [];
        const productId = items[0]?.price?.product_id || items[0]?.price?.name || "unknown";

        const { error } = await supabase
          .from("user_entitlements")
          .upsert({
            user_id: userId,
            paddle_customer_id: customerId,
            plan_type: productId,
          }, { onConflict: "user_id" });

        if (error) console.error("Transaction upsert error:", error);
        break;
      }

      case "subscription.activated":
      case "subscription.updated": {
        const subscriptionId = data.id;
        const items = data.items || [];
        const productId = items[0]?.price?.product_id || items[0]?.price?.name || "unknown";

        const { error } = await supabase
          .from("user_entitlements")
          .upsert({
            user_id: userId,
            paddle_customer_id: customerId,
            paddle_subscription_id: subscriptionId,
            plan_type: productId,
          }, { onConflict: "user_id" });

        if (error) console.error("Subscription upsert error:", error);
        break;
      }

      case "subscription.canceled":
      case "subscription.past_due": {
        const subscriptionId = data.id;

        const { error } = await supabase
          .from("user_entitlements")
          .update({
            paddle_subscription_id: null,
            plan_type: "inactive" // or free based on your app logic
          })
          .eq("user_id", userId)
          .eq("paddle_subscription_id", subscriptionId);

        if (error) console.error("Subscription update error:", error);
        break;
      }

      default:
        // Ignore unhandled event types
        break;
    }

    return new Response("OK", { status: 200 });

  } catch (error) {
    console.error("Error processing webhook:", error);
    return new Response("Internal Server Error", { status: 500 });
  }
});
