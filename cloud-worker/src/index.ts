export interface Env {
	KV_CACHE: KVNamespace;
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

export default {
	async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
		if (request.method === 'OPTIONS') {
			return handleOptions(request);
		}

		const url = new URL(request.url);
		const path = url.pathname;
		const method = request.method;

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
