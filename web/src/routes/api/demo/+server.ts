import { env } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';
import { createClient } from '@supabase/supabase-js';
import { json } from '@sveltejs/kit';
import crypto from 'crypto';

export async function POST({ request }) {
    try {
        const body = await request.json();
        const code = body.code;

        if (!code) {
            return json({ error: 'Referral code is required' }, { status: 400 });
        }

        const supabaseUrl = publicEnv.PUBLIC_SUPABASE_URL as string;
        const supabaseServiceKey = env.SUPABASE_SERVICE_ROLE_KEY as string;

        // Note: In some setups, we might just fail gracefully if Supabase isn't configured,
        // but we assume it's set up for the referral mode to work.
        if (supabaseUrl && supabaseServiceKey) {
            const supabase = createClient(supabaseUrl, supabaseServiceKey);

            // Log the referral code click
            const { error: dbError } = await supabase
                .from('referrals')
                .insert([{
                    referrer_id: code,
                    status: 'clicked'
                }]);

            if (dbError) {
                console.error("Failed to insert referral record:", dbError);
                // We'll still allow the demo even if tracking fails, as it's a demo
            }
        } else {
             console.warn("Supabase credentials missing. Referral not tracked.");
        }

        // Generate a simple token and expiration (15 minutes from now)
        const tokenBytes = crypto.randomBytes(32).toString('hex');
        const token = `demo_${tokenBytes}`;

        const now = new Date();
        const expiresAt = new Date(now.getTime() + 15 * 60 * 1000); // +15 mins

        return json({
            token,
            expiresAt: expiresAt.toISOString(),
            status: 'active'
        });

    } catch (err: any) {
        console.error("Demo API Error:", err);
        return json({ error: 'Internal Server Error' }, { status: 500 });
    }
}
