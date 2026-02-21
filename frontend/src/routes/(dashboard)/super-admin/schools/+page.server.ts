import { redirect, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { platform } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.user?.role !== 'super_admin') {
		throw redirect(302, '/');
	}

	const data = await platform.listSchools(locals.sessionToken!);
	return { schools: data.schools, total: data.total };
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		if (locals.user?.role !== 'super_admin') {
			return fail(403, { error: 'Forbidden' });
		}

		const formData = await request.formData();
		const name = formData.get('name') as string;
		const address = formData.get('address') as string;

		if (!name || name.trim().length === 0) {
			return fail(400, { error: 'School name is required' });
		}

		try {
			const result = await platform.createSchool(
				{ name: name.trim(), address: address?.trim() || undefined },
				locals.sessionToken!
			);
			return { success: true, school_id: result.school_id };
		} catch (err) {
			return fail(500, { error: 'Failed to create school' });
		}
	}
};
