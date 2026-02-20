import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/api';

export const load: PageServerLoad = async ({ locals, url }) => {
	// Redirect already-authenticated users to their dashboard.
	if (locals.user?.mfaDone) {
		const role = locals.user.role;
		throw redirect(302, `/${role}`);
	}
	return { registered: url.searchParams.has('registered') };
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
			const { data: result, setCookie } = await auth.login(email, password);

			// Forward the session cookie from the Go API to the browser.
			// apiFetch runs server-side; the Go API's Set-Cookie header is
			// silently discarded by the fetch runtime unless we re-set it here.
			if (setCookie) {
				const parts = setCookie.split(';').map((s) => s.trim());
				const eqIdx = parts[0].indexOf('=');
				const cookieName = parts[0].slice(0, eqIdx);
				const cookieValue = parts[0].slice(eqIdx + 1);
				const maxAgeStr = parts
					.find((p) => p.toLowerCase().startsWith('max-age='))
					?.split('=')[1];
				cookies.set(cookieName, cookieValue, {
					path: '/',
					httpOnly: true,
					secure: true,
					sameSite: 'strict',
					...(maxAgeStr ? { maxAge: parseInt(maxAgeStr, 10) } : {})
				});
			}

			if (result.mfa_required) {
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
