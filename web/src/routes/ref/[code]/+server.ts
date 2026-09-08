import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async ({ params }) => {
    const code = params.code;

    if (code) {
        throw redirect(302, `/demo?code=${code}`);
    }

    throw redirect(302, '/demo');
};
