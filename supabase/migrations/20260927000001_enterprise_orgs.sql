-- ============================================================
-- Phase 85: Enterprise Seat Licensing Schema
-- ============================================================

-- 1. Organizations table: one row per company that bought a license
CREATE TABLE IF NOT EXISTS public.organizations (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name            text NOT NULL,                          -- "Acme Corp"
  email_domain    text NOT NULL UNIQUE,                   -- "acme.com"
  max_seats       integer NOT NULL DEFAULT 1,             -- seats purchased
  status          text NOT NULL DEFAULT 'active'          -- 'active' | 'suspended' | 'expired'
                  CHECK (status IN ('active','suspended','expired')),
  plan_type       text NOT NULL DEFAULT 'enterprise',     -- for forward-compat
  created_at      timestamptz NOT NULL DEFAULT now(),
  updated_at      timestamptz NOT NULL DEFAULT now()
);

-- Row-level security: only service-role can write; anon can read their own domain
ALTER TABLE public.organizations ENABLE ROW LEVEL SECURITY;

CREATE POLICY "org_read_by_domain" ON public.organizations
  FOR SELECT USING (true); -- reads filtered by edge function, not RLS

-- 2. Extend `user_entitlements` with enterprise columns (additive, all nullable)
ALTER TABLE public.user_entitlements
  ADD COLUMN IF NOT EXISTS org_id          uuid REFERENCES public.organizations(id),
  ADD COLUMN IF NOT EXISTS is_enterprise   boolean NOT NULL DEFAULT false,
  ADD COLUMN IF NOT EXISTS seat_claimed_at timestamptz;

-- 3. Index for fast seat-count lookups
CREATE INDEX IF NOT EXISTS idx_user_entitlements_org_id
  ON public.user_entitlements (org_id)
  WHERE is_enterprise = true;

-- 4. Helper view: active seat count per org
CREATE OR REPLACE VIEW public.org_seat_usage AS
  SELECT
    org_id,
    COUNT(*) AS seats_used
  FROM public.user_entitlements
  WHERE is_enterprise = true
    AND plan_type = 'enterprise'
  GROUP BY org_id;
