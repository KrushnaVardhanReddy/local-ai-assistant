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

		let userId = '';

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
									userId = data[0].user_id;
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
		} else if (method === 'POST' && path === '/api/ask') {
			try {
				const body: any = await request.json();
				const embedding: number[] = body.embedding;

				if (!embedding) {
					response = Response.json({ status: 'error', message: 'Missing embedding' }, { status: 400 });
				} else if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
					const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, '');
					const rpcUrl = `${supabaseUrl}/rest/v1/rpc/match_qa_cache`;

					const rpcRes = await fetch(rpcUrl, {
						method: 'POST',
						headers: {
							'apikey': env.SUPABASE_SERVICE_ROLE_KEY,
							'Authorization': `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
							'Content-Type': 'application/json'
						},
						body: JSON.stringify({
							query_embedding: embedding,
							match_threshold: 0.92,
							match_count: 1,
							p_user_id: userId
						})
					});

					if (rpcRes.ok) {
						const matches: any = await rpcRes.json();
						if (matches && matches.length > 0) {
							response = Response.json({ status: 'success', cached_answer: matches[0].answer });
						} else {
							response = Response.json({ status: 'success', cached_answer: null });
						}
					} else {
						console.error("Error calling match_qa_cache RPC", await rpcRes.text());
						response = Response.json({ status: 'error', message: 'Failed to query cache' }, { status: 500 });
					}
				} else {
					response = Response.json({ status: 'error', message: 'Supabase configuration missing' }, { status: 500 });
				}
			} catch (e: any) {
				response = Response.json({ status: 'error', message: e.message }, { status: 500 });
			}
		} else if (method === 'POST' && path === '/api/cache/prewarm') {
			try {
				const body: any = await request.json();
				const qaPairs: any[] = body.qa_pairs || [];

				if (qaPairs.length > 0 && env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
					const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, '');
					const insertUrl = `${supabaseUrl}/rest/v1/qa_cache`;

					const recordsToInsert = qaPairs.map(pair => ({
						user_id: userId,
						question: pair.question,
						answer: pair.answer,
						embedding: pair.embedding
					}));

					const insertRes = await fetch(insertUrl, {
						method: 'POST',
						headers: {
							'apikey': env.SUPABASE_SERVICE_ROLE_KEY,
							'Authorization': `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`,
							'Content-Type': 'application/json',
							'Prefer': 'return=minimal'
						},
						body: JSON.stringify(recordsToInsert)
					});

					if (!insertRes.ok) {
						console.error("Failed to insert prewarm vectors", await insertRes.text());
					}
				}

				response = Response.json({ status: 'success', stored_count: qaPairs.length });
			} catch (e: any) {
				response = Response.json({ status: 'error', message: e.message }, { status: 500 });
			}
		} else if (method === 'DELETE' && path === '/api/cache') {
			try {
				let ids: string[] = [];
				if (request.body) {
					try {
						const body: any = await request.json();
						ids = body.ids || [];
					} catch (e) {
						// Ignored, likely empty body
					}
				}

				if (env.SUPABASE_URL && env.SUPABASE_SERVICE_ROLE_KEY) {
					const supabaseUrl = env.SUPABASE_URL.replace(/\/$/, '');
					let deleteUrl = `${supabaseUrl}/rest/v1/qa_cache?user_id=eq.${userId}`;

					if (ids.length > 0) {
						deleteUrl += `&id=in.(${ids.join(',')})`;
					}

					const deleteRes = await fetch(deleteUrl, {
						method: 'DELETE',
						headers: {
							'apikey': env.SUPABASE_SERVICE_ROLE_KEY,
							'Authorization': `Bearer ${env.SUPABASE_SERVICE_ROLE_KEY}`
						}
					});

					if (!deleteRes.ok) {
						console.error("Failed to delete vectors", await deleteRes.text());
					}
				}

				response = Response.json({ status: 'success', message: 'Cache cleared' });
			} catch (e: any) {
				response = Response.json({ status: 'error', message: e.message }, { status: 500 });
			}
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
