CREATE OR REPLACE FUNCTION increment_payg_sessions(p_user_id UUID, p_amount INT)
RETURNS VOID
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
BEGIN
    UPDATE profiles
    SET payg_sessions = COALESCE(payg_sessions, 0) + p_amount
    WHERE id = p_user_id;
END;
$$;
