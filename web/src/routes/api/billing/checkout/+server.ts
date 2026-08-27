import { env } from '$env/dynamic/private';
import Stripe from 'stripe';
import { json } from '@sveltejs/kit';

export async function POST({ locals, url }) {
    if (!locals.session || !locals.user) return new Response('Unauthorized', { status: 401 });

    const stripe = new Stripe(env.STRIPE_SECRET_KEY as string, { apiVersion: '2026-08-26.dahlia' });

    const userId = locals.user.id;
    const email = locals.user.email;

    const { data: profile } = await locals.supabase.from('profiles').select('stripe_customer_id').eq('id', userId).single();
    let stripeCustomerId = profile?.stripe_customer_id;

    if (!stripeCustomerId) {
        const customer = await stripe.customers.create({ email });
        stripeCustomerId = customer.id;
        await locals.supabase.from('profiles').update({ stripe_customer_id: stripeCustomerId }).eq('id', userId);
    }

    const origin = url.origin;
    const session = await stripe.checkout.sessions.create({
        customer: stripeCustomerId,
        mode: 'subscription',
        line_items: [{ price: env.STRIPE_PRO_PRICE_ID as string, quantity: 1 }],
        success_url: `${origin}/dashboard/billing?success=true`,
        cancel_url: `${origin}/pricing`,
        metadata: { supabase_user_id: userId }
    });

    return json({ url: session.url });
}
