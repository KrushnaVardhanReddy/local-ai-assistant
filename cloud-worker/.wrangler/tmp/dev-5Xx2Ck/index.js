var __defProp = Object.defineProperty;
var __name = (target, value) => __defProp(target, "name", { value, configurable: true });

// .wrangler/tmp/bundle-u0SUFE/strip-cf-connecting-ip-header.js
function stripCfConnectingIPHeader(input, init) {
  const request = new Request(input, init);
  request.headers.delete("CF-Connecting-IP");
  return request;
}
__name(stripCfConnectingIPHeader, "stripCfConnectingIPHeader");
globalThis.fetch = new Proxy(globalThis.fetch, {
  apply(target, thisArg, argArray) {
    return Reflect.apply(target, thisArg, [
      stripCfConnectingIPHeader.apply(null, argArray)
    ]);
  }
});

// src/index.ts
var corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "GET, POST, DELETE, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type, Authorization"
};
function handleOptions(request) {
  if (request.headers.get("Origin") !== null && request.headers.get("Access-Control-Request-Method") !== null && request.headers.get("Access-Control-Request-Headers") !== null) {
    return new Response(null, {
      headers: corsHeaders
    });
  } else {
    return new Response(null, {
      headers: {
        Allow: "GET, POST, DELETE, OPTIONS"
      }
    });
  }
}
__name(handleOptions, "handleOptions");
async function sha256Hex(message) {
  const msgBuffer = new TextEncoder().encode(message);
  const hashBuffer = await crypto.subtle.digest("SHA-256", msgBuffer);
  const hashArray = Array.from(new Uint8Array(hashBuffer));
  return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
}
__name(sha256Hex, "sha256Hex");
var src_default = {
  async fetch(request, env, ctx) {
    if (request.method === "OPTIONS") {
      return handleOptions(request);
    }
    const url = new URL(request.url);
    const path = url.pathname;
    const method = request.method;
    const protectedRoutes = ["/api/ask", "/api/cache", "/api/status", "/create-checkout-session"];
    let isProtected = false;
    for (const route of protectedRoutes) {
      if (path === route || path.startsWith(route + "/")) {
        isProtected = true;
        break;
      }
    }
    let userId = "";
    if (isProtected) {
      const authHeader = request.headers.get("Authorization");
      let isAuthorized = false;
      if (authHeader && authHeader.startsWith("Bearer ")) {
        const token = authHeader.substring(7);
        const fullHash = await sha256Hex(token);
        const keyHash = fullHash.substring(0, 12);
        if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
          const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, "");
          const queryUrl = `${supabaseUrl}/rest/v1/user_api_keys?key_hash=eq.${encodeURIComponent(keyHash)}&select=user_id,profiles!inner(payg_sessions,plan)`;
          try {
            const supabaseRes = await fetch(queryUrl, {
              method: "GET",
              headers: {
                "apikey": env.SUPABASE_SERVICE_ROLE_KEY,
                "Authorization": `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
                "Content-Type": "application/json"
              }
            });
            if (supabaseRes.ok) {
              const data = await supabaseRes.json();
              if (data && data.length > 0) {
                const profile = data[0].profiles;
                if (profile && (profile.payg_sessions > 0 || profile.plan === "lifetime")) {
                  isAuthorized = true;
                  userId = data[0].user_id;
                }
              }
            }
          } catch (error) {
            console.error("Error validating token with Supabase", error);
          }
        }
      }
      if (!isAuthorized) {
        const unauthorizedResponse = new Response("Unauthorized", { status: 401 });
        for (const [key, value] of Object.entries(corsHeaders)) {
          unauthorizedResponse.headers.set(key, value);
        }
        return unauthorizedResponse;
      }
    }
    let response;
    if (method === "GET" && path === "/api/status") {
      response = Response.json({ status: "ok", version: "1.0.0", service: "cloud-worker" });
    } else if (method === "GET" && path === "/api/cache/stats") {
      response = Response.json({
        total_entries: 0,
        total_size_bytes: 0,
        cache_hit_rate: 0
      });
    } else if (method === "POST" && path === "/create-checkout-session") {
      try {
        const body = await request.json();
        const targetUserId = body.userId || userId;
        if (!env.STRIPE_SECRET_KEY) {
          response = Response.json({ status: "error", message: "Stripe configuration missing" }, { status: 500 });
        } else {
          const origin = request.headers.get("Origin") || "http://localhost:1420";
          const stripePayload = new URLSearchParams({
            "payment_method_types[0]": "card",
            "line_items[0][price_data][currency]": "usd",
            "line_items[0][price_data][product_data][name]": "5-Pack of Interviews",
            "line_items[0][price_data][unit_amount]": "1000",
            "line_items[0][quantity]": "1",
            "mode": "payment",
            "success_url": `${origin}/?checkout=success`,
            "cancel_url": `${origin}/?checkout=canceled`,
            "client_reference_id": targetUserId
          });
          const stripeRes = await fetch("https://api.stripe.com/v1/checkout/sessions", {
            method: "POST",
            headers: {
              "Authorization": `Bearer ${env.STRIPE_SECRET_KEY}`,
              "Content-Type": "application/x-www-form-urlencoded"
            },
            body: stripePayload.toString()
          });
          if (stripeRes.ok) {
            const stripeData = await stripeRes.json();
            response = Response.json({ url: stripeData.url });
          } else {
            console.error("Stripe error", await stripeRes.text());
            response = Response.json({ status: "error", message: "Failed to create checkout session" }, { status: 500 });
          }
        }
      } catch (e) {
        response = Response.json({ status: "error", message: e.message }, { status: 500 });
      }
    } else if (method === "POST" && path === "/api/ask") {
      try {
        const body = await request.json();
        const embedding = body.embedding;
        if (!embedding) {
          response = Response.json({ status: "error", message: "Missing embedding" }, { status: 400 });
        } else if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
          const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, "");
          const rpcUrl = `${supabaseUrl}/rest/v1/rpc/match_qa_cache`;
          const rpcRes = await fetch(rpcUrl, {
            method: "POST",
            headers: {
              "apikey": env.SUPABASE_SERVICE_ROLE_KEY,
              "Authorization": `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
              "Content-Type": "application/json"
            },
            body: JSON.stringify({
              query_embedding: embedding,
              match_threshold: 0.92,
              match_count: 1,
              p_user_id: userId
            })
          });
          if (rpcRes.ok) {
            const matches = await rpcRes.json();
            if (matches && matches.length > 0) {
              response = Response.json({ status: "success", cached_answer: matches[0].answer });
            } else {
              response = Response.json({ status: "success", cached_answer: null });
            }
          } else {
            console.error("Error calling match_qa_cache RPC", await rpcRes.text());
            response = Response.json({ status: "error", message: "Failed to query cache" }, { status: 500 });
          }
        } else {
          response = Response.json({ status: "error", message: "Supabase configuration missing" }, { status: 500 });
        }
      } catch (e) {
        response = Response.json({ status: "error", message: e.message }, { status: 500 });
      }
    } else if (method === "POST" && path === "/api/cache/prewarm") {
      try {
        const body = await request.json();
        const qaPairs = body.qa_pairs || [];
        if (qaPairs.length > 0 && env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
          const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, "");
          const insertUrl = `${supabaseUrl}/rest/v1/qa_cache`;
          const recordsToInsert = qaPairs.map((pair) => ({
            user_id: userId,
            question: pair.question,
            answer: pair.answer,
            embedding: pair.embedding
          }));
          const insertRes = await fetch(insertUrl, {
            method: "POST",
            headers: {
              "apikey": env.SUPABASE_SERVICE_ROLE_KEY,
              "Authorization": `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
              "Content-Type": "application/json",
              "Prefer": "return=minimal"
            },
            body: JSON.stringify(recordsToInsert)
          });
          if (!insertRes.ok) {
            console.error("Failed to insert prewarm vectors", await insertRes.text());
          }
        }
        response = Response.json({ status: "success", stored_count: qaPairs.length });
      } catch (e) {
        response = Response.json({ status: "error", message: e.message }, { status: 500 });
      }
    } else if (method === "DELETE" && path === "/api/cache") {
      try {
        let ids = [];
        if (request.body) {
          try {
            const body = await request.json();
            ids = body.ids || [];
          } catch (e) {
          }
        }
        if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
          const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, "");
          let deleteUrl = `${supabaseUrl}/rest/v1/qa_cache?user_id=eq.${userId}`;
          if (ids.length > 0) {
            deleteUrl += `&id=in.(${ids.join(",")})`;
          }
          const deleteRes = await fetch(deleteUrl, {
            method: "DELETE",
            headers: {
              "apikey": env.SUPABASE_SERVICE_ROLE_KEY,
              "Authorization": `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`
            }
          });
          if (!deleteRes.ok) {
            console.error("Failed to delete vectors", await deleteRes.text());
          }
        }
        response = Response.json({ status: "success", message: "Cache cleared" });
      } catch (e) {
        response = Response.json({ status: "error", message: e.message }, { status: 500 });
      }
    } else if (method === "POST" && path === "/api/resume/context") {
      response = Response.json({ status: "success", message: "Resume context updated" });
    } else if (method === "POST" && path === "/api/webhook/stripe") {
      try {
        const signature = request.headers.get("stripe-signature");
        if (!signature || !env.STRIPE_WEBHOOK_SECRET) {
          return Response.json({ status: "error", message: "Missing signature or secret" }, { status: 400 });
        }
        const body = await request.text();
        const sigParts = signature.split(",").reduce((acc, part) => {
          const [key2, value] = part.split("=");
          acc[key2] = value;
          return acc;
        }, {});
        if (!sigParts.t || !sigParts.v1) {
          return Response.json({ status: "error", message: "Invalid signature format" }, { status: 400 });
        }
        const encoder = new TextEncoder();
        const signedPayload = `${sigParts.t}.${body}`;
        const key = await crypto.subtle.importKey(
          "raw",
          encoder.encode(env.STRIPE_WEBHOOK_SECRET),
          { name: "HMAC", hash: "SHA-256" },
          false,
          ["verify"]
        );
        const signatureBytes = new Uint8Array(sigParts.v1.match(/.{1,2}/g).map((byte) => parseInt(byte, 16)));
        const isValid = await crypto.subtle.verify(
          "HMAC",
          key,
          signatureBytes,
          encoder.encode(signedPayload)
        );
        if (!isValid) {
          return Response.json({ status: "error", message: "Invalid signature" }, { status: 400 });
        }
        const payload = JSON.parse(body);
        let eventType = payload.type;
        let dataObject = payload.data?.object || {};
        if (eventType === "checkout.session.completed") {
          const customerId = dataObject.client_reference_id;
          if (customerId && env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
            const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, "");
            const rpcUrl = `${supabaseUrl}/rest/v1/rpc/increment_payg_sessions`;
            const rpcRes = await fetch(rpcUrl, {
              method: "POST",
              headers: {
                "apikey": env.SUPABASE_SERVICE_ROLE_KEY,
                "Authorization": `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
                "Content-Type": "application/json"
              },
              body: JSON.stringify({ p_user_id: customerId, p_amount: 5 })
            });
            if (!rpcRes.ok) {
              console.error("Failed to increment payg_sessions via RPC", await rpcRes.text());
            }
          }
        }
        response = Response.json({ status: "success" });
      } catch (e) {
        console.error("Webhook error", e);
        response = Response.json({ status: "error", message: e.message }, { status: 400 });
      }
    } else {
      response = new Response("Not Found", { status: 404 });
    }
    for (const [key, value] of Object.entries(corsHeaders)) {
      response.headers.set(key, value);
    }
    return response;
  }
};

// node_modules/wrangler/templates/middleware/middleware-ensure-req-body-drained.ts
var drainBody = /* @__PURE__ */ __name(async (request, env, _ctx, middlewareCtx) => {
  try {
    return await middlewareCtx.next(request, env);
  } finally {
    try {
      if (request.body !== null && !request.bodyUsed) {
        const reader = request.body.getReader();
        while (!(await reader.read()).done) {
        }
      }
    } catch (e) {
      console.error("Failed to drain the unused request body.", e);
    }
  }
}, "drainBody");
var middleware_ensure_req_body_drained_default = drainBody;

// node_modules/wrangler/templates/middleware/middleware-miniflare3-json-error.ts
function reduceError(e) {
  return {
    name: e?.name,
    message: e?.message ?? String(e),
    stack: e?.stack,
    cause: e?.cause === void 0 ? void 0 : reduceError(e.cause)
  };
}
__name(reduceError, "reduceError");
var jsonError = /* @__PURE__ */ __name(async (request, env, _ctx, middlewareCtx) => {
  try {
    return await middlewareCtx.next(request, env);
  } catch (e) {
    const error = reduceError(e);
    return Response.json(error, {
      status: 500,
      headers: { "MF-Experimental-Error-Stack": "true" }
    });
  }
}, "jsonError");
var middleware_miniflare3_json_error_default = jsonError;

// .wrangler/tmp/bundle-u0SUFE/middleware-insertion-facade.js
var __INTERNAL_WRANGLER_MIDDLEWARE__ = [
  middleware_ensure_req_body_drained_default,
  middleware_miniflare3_json_error_default
];
var middleware_insertion_facade_default = src_default;

// node_modules/wrangler/templates/middleware/common.ts
var __facade_middleware__ = [];
function __facade_register__(...args) {
  __facade_middleware__.push(...args.flat());
}
__name(__facade_register__, "__facade_register__");
function __facade_invokeChain__(request, env, ctx, dispatch, middlewareChain) {
  const [head, ...tail] = middlewareChain;
  const middlewareCtx = {
    dispatch,
    next(newRequest, newEnv) {
      return __facade_invokeChain__(newRequest, newEnv, ctx, dispatch, tail);
    }
  };
  return head(request, env, ctx, middlewareCtx);
}
__name(__facade_invokeChain__, "__facade_invokeChain__");
function __facade_invoke__(request, env, ctx, dispatch, finalMiddleware) {
  return __facade_invokeChain__(request, env, ctx, dispatch, [
    ...__facade_middleware__,
    finalMiddleware
  ]);
}
__name(__facade_invoke__, "__facade_invoke__");

// .wrangler/tmp/bundle-u0SUFE/middleware-loader.entry.ts
var __Facade_ScheduledController__ = class {
  constructor(scheduledTime, cron, noRetry) {
    this.scheduledTime = scheduledTime;
    this.cron = cron;
    this.#noRetry = noRetry;
  }
  #noRetry;
  noRetry() {
    if (!(this instanceof __Facade_ScheduledController__)) {
      throw new TypeError("Illegal invocation");
    }
    this.#noRetry();
  }
};
__name(__Facade_ScheduledController__, "__Facade_ScheduledController__");
function wrapExportedHandler(worker) {
  if (__INTERNAL_WRANGLER_MIDDLEWARE__ === void 0 || __INTERNAL_WRANGLER_MIDDLEWARE__.length === 0) {
    return worker;
  }
  for (const middleware of __INTERNAL_WRANGLER_MIDDLEWARE__) {
    __facade_register__(middleware);
  }
  const fetchDispatcher = /* @__PURE__ */ __name(function(request, env, ctx) {
    if (worker.fetch === void 0) {
      throw new Error("Handler does not export a fetch() function.");
    }
    return worker.fetch(request, env, ctx);
  }, "fetchDispatcher");
  return {
    ...worker,
    fetch(request, env, ctx) {
      const dispatcher = /* @__PURE__ */ __name(function(type, init) {
        if (type === "scheduled" && worker.scheduled !== void 0) {
          const controller = new __Facade_ScheduledController__(
            Date.now(),
            init.cron ?? "",
            () => {
            }
          );
          return worker.scheduled(controller, env, ctx);
        }
      }, "dispatcher");
      return __facade_invoke__(request, env, ctx, dispatcher, fetchDispatcher);
    }
  };
}
__name(wrapExportedHandler, "wrapExportedHandler");
function wrapWorkerEntrypoint(klass) {
  if (__INTERNAL_WRANGLER_MIDDLEWARE__ === void 0 || __INTERNAL_WRANGLER_MIDDLEWARE__.length === 0) {
    return klass;
  }
  for (const middleware of __INTERNAL_WRANGLER_MIDDLEWARE__) {
    __facade_register__(middleware);
  }
  return class extends klass {
    #fetchDispatcher = (request, env, ctx) => {
      this.env = env;
      this.ctx = ctx;
      if (super.fetch === void 0) {
        throw new Error("Entrypoint class does not define a fetch() function.");
      }
      return super.fetch(request);
    };
    #dispatcher = (type, init) => {
      if (type === "scheduled" && super.scheduled !== void 0) {
        const controller = new __Facade_ScheduledController__(
          Date.now(),
          init.cron ?? "",
          () => {
          }
        );
        return super.scheduled(controller);
      }
    };
    fetch(request) {
      return __facade_invoke__(
        request,
        this.env,
        this.ctx,
        this.#dispatcher,
        this.#fetchDispatcher
      );
    }
  };
}
__name(wrapWorkerEntrypoint, "wrapWorkerEntrypoint");
var WRAPPED_ENTRY;
if (typeof middleware_insertion_facade_default === "object") {
  WRAPPED_ENTRY = wrapExportedHandler(middleware_insertion_facade_default);
} else if (typeof middleware_insertion_facade_default === "function") {
  WRAPPED_ENTRY = wrapWorkerEntrypoint(middleware_insertion_facade_default);
}
var middleware_loader_entry_default = WRAPPED_ENTRY;
export {
  __INTERNAL_WRANGLER_MIDDLEWARE__,
  middleware_loader_entry_default as default
};
//# sourceMappingURL=index.js.map
