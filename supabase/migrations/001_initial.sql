-- Create tables
CREATE TABLE profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    email TEXT,
    plan TEXT DEFAULT 'free',
    stripe_customer_id TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE user_api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    key_hash TEXT NOT NULL,
    key_encrypted TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(user_id, provider)
);

CREATE TABLE usage_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES profiles(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    tokens_in INT NOT NULL DEFAULT 0,
    tokens_out INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- Enable Row Level Security
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_api_keys ENABLE ROW LEVEL SECURITY;
ALTER TABLE usage_events ENABLE ROW LEVEL SECURITY;

-- Create Policies
CREATE POLICY "Users can view own profile" ON profiles
    FOR SELECT USING (auth.uid() = id);

CREATE POLICY "Users can update own profile" ON profiles
    FOR UPDATE USING (auth.uid() = id);

CREATE POLICY "Users can view own api keys" ON user_api_keys
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own api keys" ON user_api_keys
    FOR INSERT WITH CHECK (auth.uid() = user_id);

CREATE POLICY "Users can update own api keys" ON user_api_keys
    FOR UPDATE USING (auth.uid() = user_id);

CREATE POLICY "Users can delete own api keys" ON user_api_keys
    FOR DELETE USING (auth.uid() = user_id);

CREATE POLICY "Users can view own usage events" ON usage_events
    FOR SELECT USING (auth.uid() = user_id);

CREATE POLICY "Users can insert own usage events" ON usage_events
    FOR INSERT WITH CHECK (auth.uid() = user_id);


-- Create profile on signup trigger
CREATE OR REPLACE FUNCTION on_auth_user_created()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER SET search_path = public
AS $$
BEGIN
    INSERT INTO profiles (id, email)
    VALUES (new.id, new.email);
    RETURN new;
END;
$$;

CREATE TRIGGER on_auth_user_created
    AFTER INSERT ON auth.users
    FOR EACH ROW EXECUTE FUNCTION on_auth_user_created();
