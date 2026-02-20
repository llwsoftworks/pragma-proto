import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	// Redirect already-authenticated users to their dashboard.
	if (locals.user?.mfaDone) {
		const role = locals.user.role;
		throw redirect(302, `/${role}`);
	}
};

export const actions: Actions = {
	default: async ({ request, cookies }) => {
		const data = await request.formData();
		const email = data.get('email')?.toString() ?? '';
		const password = data.get('password')?.toString() ?? '';

		if (!email || !password) {
			return fail(400, { error: 'Email and password are required', email });
		}

		try {
			const result = await auth.login(email, password);

			if (result.mfa_required) {
				// Redirect to MFA page — the partial JWT is already in the cookie
				// set by the Go API response (forwarded by SvelteKit).
				throw redirect(302, '/login/mfa');
			}

			// Full login success.
			const role = result.user?.role ?? 'student';
			throw redirect(302, `/${role}`);
		} catch (err: unknown) {
			if (err instanceof Response || (err as { status?: number })?.status === 302) {
				throw err; // re-throw redirects
			}
			const apiErr = err as { code?: string; message?: string };
			return fail(401, {
				error: apiErr.message ?? 'Invalid email or password',
				email
			});
		}
	}
};
