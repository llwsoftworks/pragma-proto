import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { platform } from '$lib/api';

export const load: PageServerLoad = async ({ locals, params }) => {
	if (locals.user?.role !== 'super_admin') {
		throw redirect(302, '/');
	}

	const [schoolData, usersData] = await Promise.all([
		platform.getSchool(params.schoolId, locals.sessionToken!),
		platform.listSchoolUsers(params.schoolId, locals.sessionToken!)
	]);

	return { ...schoolData, users: usersData.users };
};

export const actions: Actions = {
	updateSchool: async ({ request, locals, params }) => {
		if (locals.user?.role !== 'super_admin') {
			return fail(403, { error: 'Forbidden' });
		}

		const formData = await request.formData();
		const name = formData.get('name') as string;
		const address = formData.get('address') as string;

		try {
			await platform.updateSchool(
				params.schoolId,
				{ name: name?.trim() || undefined, address: address?.trim() || undefined },
				locals.sessionToken!
			);
			return { success: true, action: 'update' };
		} catch {
			return fail(500, { error: 'Failed to update school' });
		}
	},

	createUser: async ({ request, locals, params }) => {
		if (locals.user?.role !== 'super_admin') {
			return fail(403, { error: 'Forbidden' });
		}

		const formData = await request.formData();
		const role = formData.get('role') as string;
		const email = formData.get('email') as string;
		const password = formData.get('password') as string;
		const firstName = formData.get('first_name') as string;
		const lastName = formData.get('last_name') as string;

		if (!email || !password || !firstName || !lastName || !role) {
			return fail(400, { error: 'All fields are required' });
		}

		try {
			await platform.createSchoolUser(
				params.schoolId,
				{
					role,
					email: email.trim(),
					password,
					first_name: firstName.trim(),
					last_name: lastName.trim()
				},
				locals.sessionToken!
			);
			return { success: true, action: 'createUser' };
		} catch (err: any) {
			const msg = err?.message || 'Failed to create user';
			return fail(err?.status === 409 ? 409 : 500, { error: msg });
		}
	},

	deactivateSchool: async ({ locals, params }) => {
		if (locals.user?.role !== 'super_admin') {
			return fail(403, { error: 'Forbidden' });
		}

		try {
			const result = await platform.deleteSchool(params.schoolId, locals.sessionToken!);
			return { success: true, action: 'deactivate', users_deactivated: result.users_deactivated };
		} catch {
			return fail(500, { error: 'Failed to deactivate school' });
		}
	}
};
