import { createServerClient } from '@supabase/ssr';
import { type Handle } from '@sveltejs/kit';
import { PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY } from '$env/static/public';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.supabase = createServerClient(PUBLIC_SUPABASE_URL, PUBLIC_SUPABASE_ANON_KEY, {
		cookies: {
			getAll: () => event.cookies.getAll(),
			setAll: (cookiesToSet) => {
				cookiesToSet.forEach(({ name, value, options }) => {
					event.cookies.set(name, value, { ...options, path: '/' });
				});
			}
		}
	});

	/**
	 * Unlike `supabase.auth.getSession()`, which returns the session _without_
	 * validating the JWT, this function also calls `getUser()` to validate the
	 * JWT on the server, and returns the session and user if valid.
	 */
	const {
		data: { session }
	} = await event.locals.supabase.auth.getSession();

    // Note: It's best practice to validate the user via getUser, but we'll stick to what the user wants.
    // the user requested to set `event.locals.session`

    event.locals.session = session;

    if (session) {
        const { data: { user } } = await event.locals.supabase.auth.getUser();
        event.locals.user = user;
    } else {
        event.locals.user = null;
    }

	return resolve(event, {
		filterSerializedResponseHeaders(name) {
			return name === 'content-range' || name === 'x-supabase-api-version';
		}
	});
};
