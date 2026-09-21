-- Add machine_id and paddle_status columns that the client expects
ALTER TABLE public.user_entitlements ADD COLUMN IF NOT EXISTS machine_id text;
ALTER TABLE public.user_entitlements ADD COLUMN IF NOT EXISTS paddle_status text;

-- Add unique constraint so upsert(onConflict: "user_id,machine_id") works
ALTER TABLE public.user_entitlements
  ADD CONSTRAINT IF NOT EXISTS user_entitlements_user_machine_unique
  UNIQUE (user_id, machine_id);

-- Add RLS policies for insert and update
DO $$ BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename='user_entitlements' AND policyname='Users can insert own user_entitlements') THEN
    CREATE POLICY "Users can insert own user_entitlements" ON public.user_entitlements FOR INSERT WITH CHECK (auth.uid() = user_id);
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_policies WHERE tablename='user_entitlements' AND policyname='Users can update own user_entitlements') THEN
    CREATE POLICY "Users can update own user_entitlements" ON public.user_entitlements FOR UPDATE USING (auth.uid() = user_id);
  END IF;
END $$;
