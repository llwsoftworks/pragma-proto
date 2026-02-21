import type { PageServerLoad } from './$types';
import { courses } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	try {
		const data = await courses.mine(locals.sessionToken!);
		return { courses: data.courses ?? [] };
	} catch {
		return { courses: [] };
	}
};
