import type { PageServerLoad } from './$types';
import { dashboard } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	try {
		const data = await dashboard.get(locals.sessionToken!);
		return { dashboard: data };
	} catch {
		return { dashboard: { role: 'parent', children: [] } };
	}
};
