import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { dashboard, platform } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	if (locals.user?.role !== 'super_admin') {
		throw redirect(302, '/');
	}

	const [dashboardData, schoolsData] = await Promise.all([
		dashboard.get(locals.sessionToken!),
		platform.listSchools(locals.sessionToken!)
	]);

	return { dashboard: dashboardData, schools: schoolsData.schools, totalSchools: schoolsData.total };
};
