import { serve } from "https://deno.land/std@0.168.0/http/server.ts";
import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, x-client-info, apikey, content-type",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
};

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

    // Validate the token and fetch the user
    const supabaseUrl = Deno.env.get("SUPABASE_URL");
    const supabaseKey = Deno.env.get("SUPABASE_ANON_KEY");
    const supabaseServiceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");

    if (!supabaseUrl || !supabaseKey || !supabaseServiceKey) {
       console.error("Missing Supabase environment variables.");
       return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
    }

    const supabase = createClient(supabaseUrl, supabaseKey);
    const { data: { user }, error: userError } = await supabase.auth.getUser(token);

    if (userError || !user) {
      return new Response("Unauthorized", { status: 401, headers: corsHeaders });
    }

    // Verify entitlements
    const supabaseService = createClient(supabaseUrl, supabaseServiceKey);
    const { data: entitlements, error: entError } = await supabaseService
      .from("user_entitlements")
      .select("demo_expires_at")
      .eq("user_id", user.id)
      .order("demo_expires_at", { ascending: false })
      .limit(1)
      .single();

    if (entError || !entitlements || !entitlements.demo_expires_at) {
      return new Response("No demo entitlement found", { status: 403, headers: corsHeaders });
    }

    const expiresAt = new Date(entitlements.demo_expires_at);
    if (expiresAt <= new Date()) {
      return new Response("Demo expired. Payment required.", { status: 402, headers: corsHeaders });
    }

    // Proxy to OpenAI
    const openAiApiKey = Deno.env.get("OPENAI_API_KEY");
    if (!openAiApiKey) {
        console.error("Missing OPENAI_API_KEY environment variable.");
        return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
    }

    const body = await req.text();

    const openAiReq = new Request("https://api.openai.com/v1/chat/completions", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Authorization": `Bearer ${openAiApiKey}`,
      },
      body: body,
    });

    const openAiRes = await fetch(openAiReq);

    // Stream the response back
    const responseHeaders = new Headers(openAiRes.headers);
    responseHeaders.set("Access-Control-Allow-Origin", "*");

    return new Response(openAiRes.body, {
      status: openAiRes.status,
      headers: responseHeaders,
    });

  } catch (error) {
    console.error("Proxy error:", error);
    return new Response("Internal Server Error", { status: 500, headers: corsHeaders });
  }
});
