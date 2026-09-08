import { env } from '$env/dynamic/private';
import Stripe from 'stripe';
import { json } from '@sveltejs/kit';

export async function POST({ request, url }) {
    const body = await request.text();
    const sig = request.headers.get('stripe-signature') || '';

    // Use a newer API version that works with our types, matching checkout
    const stripe = new Stripe(env.STRIPE_SECRET_KEY as string, { apiVersion: '2022-11-15' as any });
    let event;

    try {
        event = stripe.webhooks.constructEvent(body, sig, env.STRIPE_WEBHOOK_SECRET as string);
    } catch (err: any) {
        return new Response(`Webhook Error: ${err.message}`, { status: 400 });
    }

    if (event.type === 'checkout.session.completed' || event.type === 'invoice.payment_succeeded') {
        let supabaseUserId = '';
        let lineItems: Stripe.ApiList<Stripe.LineItem | Stripe.InvoiceLineItem> | undefined = undefined;

        if (event.type === 'checkout.session.completed') {
            const session = event.data.object as Stripe.Checkout.Session;
            if (session.metadata?.supabase_user_id) {
                supabaseUserId = session.metadata.supabase_user_id;
            }
            if (session.id) {
                lineItems = await stripe.checkout.sessions.listLineItems(session.id);
            }
        } else if (event.type === 'invoice.payment_succeeded') {
            const invoice = event.data.object as any;
            // Fetch the subscription to get the metadata if not on invoice
            if (invoice.subscription) {
                 const subscription = await stripe.subscriptions.retrieve(invoice.subscription as string);
                 if (subscription.metadata?.supabase_user_id) {
                     supabaseUserId = subscription.metadata.supabase_user_id;
                 }
            }
            lineItems = invoice.lines;
        }

        if (supabaseUserId && lineItems && lineItems.data && lineItems.data.length > 0) {
            const firstItem = lineItems.data[0];
            const priceId = (firstItem as any).price?.id;

            if (priceId) {
                let plan = '';

                // Determine plan based on price ID mappings
                if (priceId === env.STRIPE_PRICE_PAYG) {
                    plan = 'payg';
                } else if (priceId === env.STRIPE_PRICE_MONTHLY) {
                    plan = 'monthly';
                } else if (priceId === env.STRIPE_PRICE_FOUNDING) {
                    plan = 'founding';
                }

                if (plan) {
                    // Call Python backend to execute Supabase logic in `backend/auth.py`
                    // The backend runs on port 8765 as defined in config
                    try {
                        const backendUrl = env.BACKEND_INTERNAL_URL || 'http://127.0.0.1:8765';
                        const res = await fetch(`${backendUrl}/api/internal/billing/update_plan`, {
                            method: 'POST',
                            headers: {
                                'Content-Type': 'application/json',
                                'x-internal-secret': env.SUPABASE_SERVICE_ROLE_KEY as string
                            },
                            body: JSON.stringify({
                                user_id: supabaseUserId,
                                plan: plan
                            })
                        });

                        if (!res.ok) {
                            console.error(`Backend failed to update plan: ${res.status} ${res.statusText}`);
                        }
                    } catch (err) {
                        console.error('Failed to communicate with backend:', err);
                    }
                }
            }
        }
    }

    return json({ received: true });
}
