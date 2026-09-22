-- Migration: 20261000000000_add_referral_system
-- Adds the referral engine columns to user_entitlements

-- 1. allowed_devices: how many machines this user is allowed to activate.
--    Default is 1. Incremented to 2 when the user has a successful referral.
ALTER TABLE public.user_entitlements
  ADD COLUMN IF NOT EXISTS allowed_devices INTEGER NOT NULL DEFAULT 1;

-- 2. referral_code: a short, unique, shareable code the user gives to friends.
--    Generated automatically via a DB trigger (see below). Format: BARN-XXXX.
ALTER TABLE public.user_entitlements
  ADD COLUMN IF NOT EXISTS referral_code TEXT UNIQUE;

-- 3. referred_by_code: the referral code that was used when THIS user purchased.
--    Populated by the Paddle webhook (P66-T2) when a new user completes checkout.
--    NULL = this user was not referred by anyone.
ALTER TABLE public.user_entitlements
  ADD COLUMN IF NOT EXISTS referred_by_code TEXT;

-- 4. referral_rewarded_at: timestamp of when the referrer received their reward.
--    NULL = reward not yet granted. Prevents double-granting on webhook retries.
ALTER TABLE public.user_entitlements
  ADD COLUMN IF NOT EXISTS referral_rewarded_at TIMESTAMPTZ;

-- 5. Create a DB function + trigger to auto-generate referral_code when a new row
--    is inserted and referral_code is NULL.
CREATE OR REPLACE FUNCTION generate_referral_code()
RETURNS TRIGGER AS $$
DECLARE
  code TEXT;
  collision BOOLEAN := TRUE;
BEGIN
  -- Keep generating until we find a unique 4-char suffix
  WHILE collision LOOP
    code := 'BARN-' || upper(substring(md5(random()::text) from 1 for 4));
    collision := EXISTS (
      SELECT 1 FROM public.user_entitlements WHERE referral_code = code
    );
  END LOOP;
  NEW.referral_code := code;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Only fire on INSERT when referral_code is NULL (won't re-generate on updates)
CREATE OR REPLACE TRIGGER trg_generate_referral_code
  BEFORE INSERT ON public.user_entitlements
  FOR EACH ROW
  WHEN (NEW.referral_code IS NULL)
  EXECUTE FUNCTION generate_referral_code();

-- 6. RLS: Users can now also read their own referral_code and allowed_devices.
--    The existing "Users can select own user_entitlements" policy already covers this.
--    No new RLS policies needed.

-- 7. Service-role can update allowed_devices (for the webhook reward logic).
--    The Edge Function uses the service role key, so it bypasses RLS. No change needed.
