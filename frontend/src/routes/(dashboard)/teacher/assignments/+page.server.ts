import type { PageServerLoad } from './$types';
import { assignments } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	try {
		const data = await assignments.list(locals.sessionToken!);
		return { assignments: data.assignments ?? [] };
	} catch {
		return { assignments: [] };
	}
};
