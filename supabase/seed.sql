-- Insert a dummy user into auth.users
-- ID: 11111111-1111-1111-1111-111111111111
INSERT INTO auth.users (
    instance_id,
    id,
    aud,
    role,
    email,
    encrypted_password,
    email_confirmed_at,
    recovery_sent_at,
    last_sign_in_at,
    raw_app_meta_data,
    raw_user_meta_data,
    created_at,
    updated_at,
    confirmation_token,
    email_change,
    email_change_token_new,
    recovery_token
) VALUES (
    '00000000-0000-0000-0000-000000000000',
    '11111111-1111-1111-1111-111111111111',
    'authenticated',
    'authenticated',
    'test@example.com',
    crypt('password123', gen_salt('bf')),
    now(),
    now(),
    now(),
    '{"provider":"email","providers":["email"]}',
    '{}',
    now(),
    now(),
    '',
    '',
    '',
    ''
) ON CONFLICT (id) DO NOTHING;

-- Update the newly created profile (created via the on_auth_user_created trigger)
-- Set payg_sessions = 10
UPDATE public.profiles
SET payg_sessions = 10, plan = 'free'
WHERE id = '11111111-1111-1111-1111-111111111111';

-- Insert a test API key record into public.user_api_keys
-- Raw key: sk_test_1234567890abcdef
-- Hash (first 12 chars of SHA-256): dc3d31697fac
INSERT INTO public.user_api_keys (
    id,
    user_id,
    provider,
    key_hash,
    key_encrypted,
    created_at
) VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'openai',
    'dc3d31697fac',
    'encrypted_dummy_value',
    now()
) ON CONFLICT (user_id, provider) DO UPDATE
SET key_hash = EXCLUDED.key_hash,
    key_encrypted = EXCLUDED.key_encrypted;

-- Note for testing:
-- Use the API key 'sk_test_1234567890abcdef' in the frontend.
