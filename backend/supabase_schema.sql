-- ─────────────────────────────────────────────────────────────────
-- Phase 20: Anti-Sharing — Device Registration & Session Lock
-- ─────────────────────────────────────────────────────────────────

-- Device registration table
-- Tracks which machines are authorized per user account.
CREATE TABLE IF NOT EXISTS user_devices (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    machine_id      TEXT NOT NULL,          -- UUID from OS keychain (lib.rs get_machine_id)
    device_label    TEXT DEFAULT 'My Device', -- User-friendly label (editable from dashboard)
    registered_at   TIMESTAMPTZ DEFAULT NOW(),
    last_seen_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, machine_id)             -- One entry per user+machine combination
);

-- Index for fast lookup by user
CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);

-- Active session tracking column on users table (via Supabase profiles table)
-- If you have a 'profiles' table, add this column. Otherwise add to auth.users metadata.
-- Run this if a profiles/users table exists:
ALTER TABLE profiles
    ADD COLUMN IF NOT EXISTS active_session_id TEXT DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS active_session_at  TIMESTAMPTZ DEFAULT NULL,
    ADD COLUMN IF NOT EXISTS demo_credits_minutes INTEGER DEFAULT 15;
    -- demo_credits_minutes: starts at 15 (free demo), increases with referrals (+15 each)

-- Device limits per plan (reference table)
CREATE TABLE IF NOT EXISTS plan_device_limits (
    plan        TEXT PRIMARY KEY,
    max_devices INTEGER NOT NULL,
    max_sessions INTEGER NOT NULL DEFAULT 1
);

INSERT INTO plan_device_limits (plan, max_devices, max_sessions) VALUES
    ('demo',     1, 1),
    ('payg',     1, 1),
    ('monthly',  2, 1),
    ('founding', 3, 1)
ON CONFLICT (plan) DO NOTHING;

-- Referrals tracking table (supports P19-T5, P19-T7)
CREATE TABLE IF NOT EXISTS referrals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    referrer_id     UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
    referee_id      UUID REFERENCES auth.users(id) ON DELETE SET NULL,
    referral_code   TEXT NOT NULL UNIQUE,
    status          TEXT DEFAULT 'pending', -- 'pending' | 'signed_up' | 'converted'
    created_at      TIMESTAMPTZ DEFAULT NOW(),
    converted_at    TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_referrals_code ON referrals(referral_code);
CREATE INDEX IF NOT EXISTS idx_referrals_referrer ON referrals(referrer_id);

-- Row Level Security (RLS)
ALTER TABLE user_devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE referrals ENABLE ROW LEVEL SECURITY;

-- Users can only see/modify their own devices
CREATE POLICY "Users manage own devices" ON user_devices
    USING (auth.uid() = user_id)
    WITH CHECK (auth.uid() = user_id);

-- Users can read their own referrals
CREATE POLICY "Users read own referrals" ON referrals
    FOR SELECT USING (auth.uid() = referrer_id OR auth.uid() = referee_id);
