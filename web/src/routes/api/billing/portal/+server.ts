import { env } from '$env/dynamic/private';
import Stripe from 'stripe';
import { json } from '@sveltejs/kit';

export async function POST({ locals, url }) {
    if (!locals.session || !locals.user) return new Response('Unauthorized', { status: 401 });

    const stripe = new Stripe(env.STRIPE_SECRET_KEY as string, { apiVersion: '2026-08-26.dahlia' });

    const userId = locals.user.id;
    const { data: profile } = await locals.supabase.from('profiles').select('stripe_customer_id').eq('id', userId).single();

    if (!profile?.stripe_customer_id) {
        return new Response('No customer ID found', { status: 400 });
    }

    const portal = await stripe.billingPortal.sessions.create({
        customer: profile.stripe_customer_id,
        return_url: `${url.origin}/dashboard/billing`
    });

    return json({ url: portal.url });
}
