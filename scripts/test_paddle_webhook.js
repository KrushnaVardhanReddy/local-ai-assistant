const crypto = require("crypto");
const fs = require("fs");
const path = require("path");

// Load secret from .env.local
const envPath = path.join(__dirname, "../.env.local");
let secret = process.env.PADDLE_WEBHOOK_SECRET;

if (!secret && fs.existsSync(envPath)) {
  const envFile = fs.readFileSync(envPath, "utf-8");
  const match = envFile.match(/PADDLE_WEBHOOK_SECRET=(.+)/);
  if (match) {
    secret = match[1].trim();
  }
}

if (!secret) {
  console.error("❌ PADDLE_WEBHOOK_SECRET not found in .env.local");
  process.exit(1);
}

// Ensure Supabase URL is correct
const WEBHOOK_URL = "http://127.0.0.1:54321/functions/v1/paddle-webhook";

const mockEvent = {
  event_id: "evt_01m30...",
  event_type: "transaction.completed",
  occurred_at: new Date().toISOString(),
  data: {
    id: "txn_01m30...",
    customer_id: "ctm_01m30...",
    custom_data: {
      user_id: "test-user-id-from-supabase"
    },
    items: [
      {
        price: {
          id: "pri_01m30...",
          product_id: "pro_barnowl_lifetime",
          name: "BarnOwl AI Lifetime"
        }
      }
    ]
  }
};

const rawBody = JSON.stringify(mockEvent);

// Generate signature (Paddle v1 format)
const ts = Math.floor(Date.now() / 1000).toString();
const signedPayload = `${ts}:${rawBody}`;
const hmac = crypto.createHmac("sha256", secret).update(signedPayload).digest("hex");
const signatureHeader = `ts=${ts};h1=${hmac}`;

async function runTest() {
  console.log(`🚀 Sending mock ${mockEvent.event_type} to ${WEBHOOK_URL}...`);
  try {
    const res = await fetch(WEBHOOK_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "paddle-signature": signatureHeader,
      },
      body: rawBody,
    });
    
    const text = await res.text();
    console.log(`✅ Response [${res.status}]:`, text);
  } catch (err) {
    console.error("❌ Request failed:", err.message);
  }
}

runTest();
