import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { auth } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	// Redirect already-authenticated users to their dashboard.
	if (locals.user?.mfaDone) {
		throw redirect(302, `/${locals.user.role}`);
	}
	return {};
};

export const actions: Actions = {
	default: async ({ request }) => {
		const data = await request.formData();
		const firstName = data.get('first_name')?.toString().trim() ?? '';
		const lastName = data.get('last_name')?.toString().trim() ?? '';
		const email = data.get('email')?.toString().trim() ?? '';
		const password = data.get('password')?.toString() ?? '';
		const confirmPassword = data.get('confirm_password')?.toString() ?? '';
		const role = data.get('role')?.toString() ?? '';
		const schoolId = data.get('school_id')?.toString().trim() ?? '';
		const phone = data.get('phone')?.toString().trim() ?? '';

		const fields = { firstName, lastName, email, role, schoolId, phone };

		if (!firstName || !lastName || !email || !password || !confirmPassword || !role || !schoolId) {
			return fail(400, { error: 'All required fields must be filled in.', ...fields });
		}

		if (password !== confirmPassword) {
			return fail(400, { error: 'Passwords do not match.', ...fields });
		}

		if (password.length < 12) {
			return fail(400, { error: 'Password must be at least 12 characters.', ...fields });
		}

		// UUID format check (client-side hint; Go API validates strictly).
		const uuidRe = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;
		if (!uuidRe.test(schoolId)) {
			return fail(400, {
				error: 'School ID must be a valid UUID (e.g. a1b2c3d4-e5f6-7890-abcd-ef1234567890).',
				...fields
			});
		}

		try {
			await auth.register(
				{
					school_id: schoolId,
					role,
					email,
					password,
					first_name: firstName,
					last_name: lastName,
					phone: phone || undefined
				},
				'' // /auth/register is a public endpoint — no token required
			);
		} catch (err: unknown) {
			const apiErr = err as { code?: string; message?: string };
			const message =
				apiErr.code === 'email_exists'
					? 'An account with this email already exists at this school.'
					: apiErr.code === 'breached_password'
						? 'This password has appeared in a known data breach. Please choose a different one.'
						: apiErr.code === 'weak_password'
							? 'Password must be at least 12 characters.'
							: apiErr.code === 'validation_error'
								? 'Please check all fields and try again.'
								: (apiErr.message ?? 'Registration failed. Please try again.');
			return fail(400, { error: message, ...fields });
		}

		throw redirect(302, '/login?registered=1');
	}
};
