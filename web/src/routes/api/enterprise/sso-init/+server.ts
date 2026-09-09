import { createClient } from '@supabase/supabase-js';
import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { env as dynamicPrivateEnv } from '$env/dynamic/private';
import { env as dynamicPublicEnv } from '$env/dynamic/public';

export const POST: RequestHandler = async ({ request }) => {
	const { domain } = await request.json();
	if (!domain) return json({ error: 'domain required' }, { status: 400 });

	const publicSupabaseUrl = dynamicPublicEnv.PUBLIC_SUPABASE_URL || process.env.PUBLIC_SUPABASE_URL;
	const supabaseServiceKey = dynamicPrivateEnv.SUPABASE_SERVICE_ROLE_KEY || process.env.SUPABASE_SERVICE_ROLE_KEY;

	if (!publicSupabaseUrl || !supabaseServiceKey) {
		return json({ error: 'Supabase configuration missing' }, { status: 500 });
	}

	const supabase = createClient(
		publicSupabaseUrl,
		supabaseServiceKey
	);

	const { data, error } = await supabase.auth.admin.generateLink({
		// @ts-ignore
		type: 'sso',
		// @ts-ignore
		options: { domain }
	});

	if (error) return json({ error: error.message }, { status: 400 });
	return json({ url: data.properties?.action_link });
};
