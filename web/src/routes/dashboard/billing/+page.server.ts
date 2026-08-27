import { redirect } from '@sveltejs/kit';

export async function load({ locals }) {
    if (!locals.session || !locals.user) throw redirect(303, '/login');

    return {
        plan: locals.user.plan || 'free',
        nextBillingDate: null
    };
}
