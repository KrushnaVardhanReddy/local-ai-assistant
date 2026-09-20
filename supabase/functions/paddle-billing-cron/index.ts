/**
 * ENV VARS required:
 * SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY
 * CRON_SECRET, PADDLE_API_KEY
 * PADDLE_RATE_MENTOR_PER_MIN, PADDLE_RATE_COUNSEL_PER_MIN, PADDLE_RATE_CLINIC_PER_MIN
 */

import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
};

function getRatePerMinute(productMode: string): number {
  switch (productMode) {
    case "mentor":
      return parseFloat(Deno.env.get("PADDLE_RATE_MENTOR_PER_MIN") || "0.02");
    case "counsel":
      return parseFloat(Deno.env.get("PADDLE_RATE_COUNSEL_PER_MIN") || "0.03");
    case "clinic":
      return parseFloat(Deno.env.get("PADDLE_RATE_CLINIC_PER_MIN") || "0.025");
    default:
      return 0;
  }
}

serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  try {
    const authHeader = req.headers.get("Authorization");
    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      return new Response("Missing or invalid Authorization header", {
        status: 401,
        headers: corsHeaders,
      });
    }

    const token = authHeader.replace("Bearer ", "");
    const cronSecret = Deno.env.get("CRON_SECRET");

    if (!cronSecret || token !== cronSecret) {
      return new Response("Unauthorized", { status: 401, headers: corsHeaders });
    }

    const supabaseUrl = Deno.env.get("SUPABASE_URL");
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");
    const paddleApiKey = Deno.env.get("PADDLE_API_KEY");

    if (!supabaseUrl || !supabaseServiceKey || !paddleApiKey) {
      console.error("Missing required environment variables");
      return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
    }

    const supabase = createClient(supabaseUrl, supabaseServiceKey);

    // 1. Query all rows from user_entitlements where conditions are met
    // billing_period_end <= now() + interval '1 day' -> meaning billing_period_end is less than or equal to tomorrow.
    // Deno doesn't support raw SQL easily through JS without rpc, so we can do it via JS filter or build an RPC.
    // Wait, Supabase JS has `.lte('billing_period_end', tomorrow.toISOString())`

    const tomorrow = new Date();
    tomorrow.setDate(tomorrow.getDate() + 1);

    // But wait, the condition is usage_seconds > included_seconds.
    // Supabase JS doesn't support column comparison in filter easily (e.g. usage_seconds > included_seconds) without RPC or Postgres function.
    // Since we just need to fetch all candidates and we can filter them in JS if there are not too many,
    // or we can just fetch all that are ending soon, and then filter in memory.

    // Better to fetch all rows ending soon and not null subscription
    const { data: candidates, error: queryError } = await supabase
      .from("user_entitlements")
      .select("*")
      .not("paddle_subscription_id", "is", null)
      .lte("billing_period_end", tomorrow.toISOString());

    if (queryError) {
      console.error("Query error:", queryError);
      return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
    }

    let charged = 0;
    let skipped = 0;
    let errors = 0;

    for (const row of candidates || []) {
      const usageSeconds = row.usage_seconds || 0;
      const includedSeconds = row.included_seconds || 0;

      if (usageSeconds <= includedSeconds) {
        // No overage. Paddle will fire subscription.updated which is the source of truth for new periods.
        // So we don't manually mutate period/usage here.
        continue;
      }

      // 2. For each row with overage:
      const overageSeconds = usageSeconds - includedSeconds;
      const overageMinutes = Math.ceil(overageSeconds / 60);
      const productMode = row.product_mode;
      const ratePerMinute = getRatePerMinute(productMode);
      const overageAmountUsd = overageMinutes * ratePerMinute;

      if (overageAmountUsd < 0.50) {
        skipped++;
        // Skip advancing period here as well? The spec says "After charging, reset usage_seconds = 0 and advance billing_period_start/end". If we don't charge, we don't mutate, and wait for Paddle to update it via webhook.
        continue;
      }

      // 3. For each chargeable row, call Paddle API
      const amountCentsStr = Math.round(overageAmountUsd * 100).toString();

      const paddleReqBody = {
        items: [{
          price: {
            description: `Usage overage - ${overageMinutes} mins`,
            unit_price: { amount: amountCentsStr, currency_code: "USD" },
            quantity: { minimum: 1, maximum: 1 }
          },
          quantity: 1
        }],
        customer_id: row.paddle_customer_id,
        collection_mode: "automatic"
      };

      try {
        const paddleRes = await fetch("https://api.paddle.com/transactions", {
          method: "POST",
          headers: {
            "Authorization": `Bearer ${paddleApiKey}`,
            "Content-Type": "application/json"
          },
          body: JSON.stringify(paddleReqBody)
        });

        if (!paddleRes.ok) {
          console.error(`Paddle API error for customer ${row.paddle_customer_id}:`, await paddleRes.text());
          errors++;
          continue; // Skip advancing period if charge failed? The spec says:
          // "After charging, reset usage_seconds = 0 and advance billing_period_start/end"
        }

        charged++;
        // 4. After charging, reset usage_seconds = 0 and advance billing_period_start/end
        await advanceBillingPeriod(supabase, row);
      } catch (e) {
        console.error("Error making request to Paddle API:", e);
        errors++;
      }
    }

    return new Response(JSON.stringify({ charged, skipped, errors }), {
      headers: { ...corsHeaders, "Content-Type": "application/json" }
    });

  } catch (error) {
    console.error("Cron error:", error);
    return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
  }
});

async function advanceBillingPeriod(supabase: any, row: any) {
    let currentStart = new Date(row.billing_period_start);
    let currentEnd = new Date(row.billing_period_end);

    if (isNaN(currentStart.getTime())) currentStart = new Date();
    if (isNaN(currentEnd.getTime())) currentEnd = new Date();

    const newStart = new Date(currentEnd);
    const newEnd = new Date(currentEnd);
    newEnd.setMonth(newEnd.getMonth() + 1);

    const { error: updateError } = await supabase
      .from("user_entitlements")
      .update({
        usage_seconds: 0,
        billing_period_start: newStart.toISOString(),
        billing_period_end: newEnd.toISOString()
      })
      .eq("user_id", row.user_id);

    if (updateError) {
      console.error("Error updating billing period for user:", row.user_id, updateError);
    }
}
