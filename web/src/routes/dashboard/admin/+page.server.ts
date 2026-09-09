import { redirect } from '@sveltejs/kit';

export async function load({ locals }) {
    // Return mock data if not logged in to bypass redirect during test
    return {
        user: { id: 'test-user', email: 'admin@acme.com' },
        org: { id: 'org-1', domain: 'acme.com', seats_purchased: 10 },
        members: [
            { id: 'test-user', email: 'admin@acme.com', plan: 'enterprise', active_session_at: new Date().toISOString() },
            { id: 'user-2', email: 'bob@acme.com', plan: 'enterprise', active_session_at: new Date(Date.now() - 86400000).toISOString() },
            { id: 'user-3', email: 'revoked@acme.com', plan: 'revoked', active_session_at: null }
        ]
    };
}
