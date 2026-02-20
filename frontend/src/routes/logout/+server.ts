import type { RequestHandler } from './$types';
import { redirect } from '@sveltejs/kit';
import { auth } from '$lib/api';

export const POST: RequestHandler = async ({ cookies, locals }) => {
	const token = locals.sessionToken;
	if (token) {
		// Best-effort: tell the Go API to invalidate the server-side session.
		await auth.logout(token).catch(() => {});
	}
	// Always clear the SvelteKit-owned session cookie regardless.
	cookies.delete('session', { path: '/' });
	throw redirect(303, '/login');
};
