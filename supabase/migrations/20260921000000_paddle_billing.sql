-- Migration: 20260921000000_paddle_billing.sql

ALTER TABLE user_entitlements
    ADD COLUMN IF NOT EXISTS paddle_subscription_id TEXT,
    ADD COLUMN IF NOT EXISTS paddle_customer_id TEXT,
    ADD COLUMN IF NOT EXISTS billing_period_start TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS billing_period_end TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS included_seconds INTEGER DEFAULT 0,
    ADD COLUMN IF NOT EXISTS license_key TEXT,
    ADD COLUMN IF NOT EXISTS product_mode TEXT,
    ADD COLUMN IF NOT EXISTS remaining_sessions INTEGER;

CREATE OR REPLACE FUNCTION get_user_id_by_email(user_email TEXT)
RETURNS UUID AS $$
DECLARE
    found_user_id UUID;
BEGIN
    SELECT id INTO found_user_id FROM auth.users WHERE email = user_email LIMIT 1;
    RETURN found_user_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
