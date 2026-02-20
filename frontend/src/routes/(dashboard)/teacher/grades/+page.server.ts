import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	// Fetch teacher's courses.
	try {
		const result = await fetch(`${process.env.API_URL ?? 'http://localhost:8080'}/courses/mine`, {
			headers: { Authorization: `Bearer ${locals.sessionToken}` }
		});
		const data = result.ok ? await result.json() : { courses: [] };
		return { courses: data.courses ?? [] };
	} catch {
		return { courses: [] };
	}
};
