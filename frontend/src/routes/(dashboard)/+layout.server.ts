import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';

/**
 * Auth guard for all dashboard routes.
 * Verifies the JWT is present and the user has completed MFA if required.
 * Loads user data from locals (set by hooks.server.ts).
 */
export const load: LayoutServerLoad = async ({ locals }) => {
	if (!locals.user) {
		throw redirect(302, '/login');
	}

	const { user } = locals;

	// MFA required roles that haven't completed MFA.
	const mfaRequiredRoles = ['super_admin', 'admin', 'teacher'];
	if (mfaRequiredRoles.includes(user.role) && !user.mfaDone) {
		throw redirect(302, '/login/mfa');
	}

	return {
		user: {
			id: user.id,
			email: user.email,
			role: user.role,
			schoolId: user.schoolId
		}
	};
};
