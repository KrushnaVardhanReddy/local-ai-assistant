export interface Env {
	KV_CACHE: KVNamespace;
	USERS_KV: KVNamespace;
	SUPABASE_URL: string;
	SUPABASE_SERVICE_ROLE_KEY: string;
}

const corsHeaders = {
	'Access-Control-Allow-Origin': '*',
	'Access-Control-Allow-Methods': 'GET, POST, DELETE, OPTIONS',
	'Access-Control-Allow-Headers': 'Content-Type, Authorization',
};

function handleOptions(request: Request) {
	if (
		request.headers.get('Origin') !== null &&
		request.headers.get('Access-Control-Request-Method') !== null &&
		request.headers.get('Access-Control-Request-Headers') !== null
	) {
		return new Response(null, {
			headers: corsHeaders,
		});
	} else {
		return new Response(null, {
			headers: {
				Allow: 'GET, POST, DELETE, OPTIONS',
			},
		});
	}
}

async function sha256Hex(message: string): Promise<string> {
	const msgBuffer = new TextEncoder().encode(message);
	const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		if (request.method === 'OPTIONS') {
			return handleOptions(request);
		}

		const url = new URL(request.url);
		const path = url.pathname;
		const method = request.method;

		// Only protect specific routes
		const protectedRoutes = ['/api/ask', '/api/cache', '/api/status'];
		let isProtected = false;
		for (const route of protectedRoutes) {
			if (path === route || path.startsWith(route + '/')) {
				isProtected = true;
				break;
			}
		}

		if (isProtected) {
			const authHeader = request.headers.get('Authorization');
			let isAuthorized = false;

			if (authHeader && authHeader.startsWith('Bearer ')) {
				const token = authHeader.substring(7);

				// Hash the token to match the key_hash stored in Supabase
				const fullHash = await sha256Hex(token);
				const keyHash = fullHash.substring(0, 12);

				// Validate against Supabase
				if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
					const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, '');
					const queryUrl = `${supabaseUrl}/rest/v1/user_api_keys?key_hash=eq.${encodeURIComponent(keyHash)}&select=user_id,profiles!inner(payg_sessions,plan)`;

					try {
						const supabaseRes = await fetch(queryUrl, {
							method: 'GET',
							headers: {
								'apikey': env.SUPABASE_SERVICE_ROLE_KEY,
								'Authorization': `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
								'Content-Type': 'application/json'
							}
						});

						if (supabaseRes.ok) {
							const data: any = await supabaseRes.json();
							if (data && data.length > 0) {
								const profile = data[0].profiles;
								if (profile && (profile.payg_sessions > 0 || profile.plan === 'lifetime')) {
									isAuthorized = true;
								}
							}
						}
					} catch (error) {
						console.error("Error validating token with Supabase", error);
					}
				}
			}

			if (!isAuthorized) {
				const unauthorizedResponse = new Response('Unauthorized', { status: 401 });
				for (const [key, value] of Object.entries(corsHeaders)) {
					unauthorizedResponse.headers.set(key, value);
				}
				return unauthorizedResponse;
			}
		}

		let response: Response;

		if (method === 'GET' && path === '/api/status') {
			response = Response.json({ status: 'ok', version: '1.0.0', service: 'cloud-worker' });
		} else if (method === 'GET' && path === '/api/cache/stats') {
			response = Response.json({
				total_entries: 0,
				total_size_bytes: 0,
				cache_hit_rate: 0.0,
			});
		} else if (method === 'DELETE' && path === '/api/cache') {
			response = Response.json({ status: 'success', message: 'Cache cleared' });
		} else if (method === 'POST' && path === '/api/resume/context') {
			response = Response.json({ status: 'success', message: 'Resume context updated' });
		} else {
			response = new Response('Not Found', { status: 404 });
		}

		// Apply CORS headers to the response
		for (const [key, value] of Object.entries(corsHeaders)) {
			response.headers.set(key, value);
		}

		return response;
	},
};
