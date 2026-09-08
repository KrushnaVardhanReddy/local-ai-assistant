-- Add payg_sessions tracking column to profiles table
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS payg_sessions INT DEFAULT 0;
