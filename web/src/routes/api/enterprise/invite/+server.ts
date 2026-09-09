import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { createClient } from '@supabase/supabase-js';

export async function POST({ request, locals }) {
    if (!locals.user) return json({ error: 'Unauthorized' }, { status: 401 });

    const supabaseAdmin = createClient(
        (env.PUBLIC_SUPABASE_URL || process.env.PUBLIC_SUPABASE_URL) as string,
        (env.SUPABASE_SERVICE_ROLE_KEY || process.env.SUPABASE_SERVICE_ROLE_KEY) as string
    );

    const { email } = await request.json();
    if (!email) return json({ error: 'email is required' }, { status: 400 });

    const { data: user } = await supabaseAdmin
        .from('users')
        .select('org_id')
        .eq('id', locals.user.id)
        .single();

    const orgId = user?.org_id;

    const { data: org } = await supabaseAdmin
        .from('enterprise_orgs')
        .select('is_org_admin')
        .eq('id', orgId)
        .single();

    if (!org || !org.is_org_admin) {
        return json({ error: 'Unauthorized, not org admin' }, { status: 403 });
    }

    if (!orgId) return json({ error: 'No org associated' }, { status: 400 });

    // invite user via auth
    const { data: authData, error: authError } = await supabaseAdmin.auth.admin.inviteUserByEmail(email);

    if (authError) {
        return json({ error: authError.message }, { status: 400 });
    }

    // Wait a brief moment to ensure trigger completes if new user
    await new Promise(resolve => setTimeout(resolve, 500));

    const { error: profileError } = await supabaseAdmin
        .from('users')
        .upsert({
            id: authData.user.id,
            email: email,
            org_id: orgId,
            plan: 'enterprise'
        });

    if (profileError) {
        return json({ error: profileError.message }, { status: 400 });
    }

    return json({ success: true });
}
