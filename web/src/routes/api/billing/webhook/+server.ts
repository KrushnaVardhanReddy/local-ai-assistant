import { env } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';
import { createClient } from '@supabase/supabase-js';
import Stripe from 'stripe';
import { json } from '@sveltejs/kit';

const processedEvents = new Set<string>();

export async function POST({ request }) {
    const body = await request.text();
    const sig = request.headers.get('stripe-signature') || '';

    const stripe = new Stripe(env.STRIPE_SECRET_KEY as string, { apiVersion: '2026-08-26.dahlia' });
    let event;

    try {
        event = stripe.webhooks.constructEvent(body, sig, env.STRIPE_WEBHOOK_SECRET as string);
    } catch (err: any) {
        return new Response(`Webhook Error: ${err.message}`, { status: 400 });
    }

    if (processedEvents.has(event.id)) {
        return json({ received: true });
    }

    const supabase = createClient(publicEnv.PUBLIC_SUPABASE_URL as string, env.SUPABASE_SERVICE_ROLE_KEY as string);

    if (event.type === 'checkout.session.completed') {
        const session = event.data.object as any;
        await supabase.from('profiles').update({ plan: 'pro' }).eq('stripe_customer_id', session.customer);
    } else if (event.type === 'customer.subscription.deleted' || event.type === 'invoice.payment_failed') {
        const obj = event.data.object as any;
        await supabase.from('profiles').update({ plan: 'free' }).eq('stripe_customer_id', obj.customer);
    }

    processedEvents.add(event.id);

    return json({ received: true });
}
