#!/bin/bash
# Run this script using: chmod +x scripts/deploy_supabase.sh && ./scripts/deploy_supabase.sh

echo "Starting Supabase Deployment Automation..."

echo "1. Linking Supabase Project..."
echo "Please enter your Supabase Project ID when prompted if not automatically detected."
supabase link

echo "2. Pushing migrations to production..."
supabase db push

echo "3. Deploying Edge Functions..."
supabase functions deploy llm-proxy

echo "4. Setting up Edge Function secrets..."
echo "To set real API keys, run: supabase secrets set OPENAI_API_KEY=<your_key> STRIPE_SECRET_KEY=<your_key>"
# A dummy prompt or instruction, or setting an example secret. The spec says: "Sets secrets (supabase secrets set OPENAI_API_KEY=...)"
supabase secrets set OPENAI_API_KEY=YOUR_OPENAI_API_KEY

echo "Deployment complete!"
