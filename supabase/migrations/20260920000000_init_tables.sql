-- Migration: 20260920000000_init_tables
-- Create dev_allowlist table
CREATE TABLE dev_allowlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    machine_id TEXT UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Create user_entitlements table
CREATE TABLE user_entitlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES auth.users(id) ON DELETE CASCADE,
    usage_seconds INTEGER DEFAULT 0,
    plan_type TEXT DEFAULT 'free',
    stripe_subscription_status TEXT,
    demo_expires_at TIMESTAMPTZ
);

-- Enable RLS
ALTER TABLE dev_allowlist ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_entitlements ENABLE ROW LEVEL SECURITY;

-- Policies for dev_allowlist
CREATE POLICY "Public can select dev_allowlist"
    ON dev_allowlist FOR SELECT USING (true);

-- Policies for user_entitlements
CREATE POLICY "Users can select own user_entitlements"
    ON user_entitlements FOR SELECT USING (auth.uid() = user_id);
