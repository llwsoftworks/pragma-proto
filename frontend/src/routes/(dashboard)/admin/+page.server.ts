import type { PageServerLoad } from './$types';
import { dashboard } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	const data = await dashboard.get(locals.sessionToken!);
	return { dashboard: data };
};
