import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { createClient } from '@supabase/supabase-js';

export async function POST({ request, locals }) {
    if (!locals.user) return json({ error: 'Unauthorized' }, { status: 401 });

    const supabaseAdmin = createClient(
        (env.PUBLIC_SUPABASE_URL || process.env.PUBLIC_SUPABASE_URL) as string,
        (env.SUPABASE_SERVICE_ROLE_KEY || process.env.SUPABASE_SERVICE_ROLE_KEY) as string
    );

    const { userId } = await request.json();
    if (!userId) return json({ error: 'userId is required' }, { status: 400 });

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

    const { data: targetUser } = await supabaseAdmin
        .from('users')
        .select('org_id')
        .eq('id', userId)
        .single();

    if (!targetUser || targetUser.org_id !== orgId) {
        return json({ error: 'User not in your org' }, { status: 403 });
    }

    const { error: updateError } = await supabaseAdmin
        .from('users')
        .update({ plan: 'revoked' })
        .eq('id', userId);

    if (updateError) {
        return json({ error: updateError.message }, { status: 400 });
    }

    return json({ success: true });
}
