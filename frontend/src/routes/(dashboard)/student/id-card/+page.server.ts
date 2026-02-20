import type { PageServerLoad } from './$types';
import { students } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

	try {
		// Resolve the student's DB id from the JWT user_id via /students/me.
		const student = await students.me(locals.sessionToken!);

		const idRes = await fetch(`${API_BASE}/students/${student.id}/digital-id`, {
			headers: { Authorization: `Bearer ${locals.sessionToken}` }
		});

		if (!idRes.ok) {
			return { digitalId: null, error: 'No digital ID found. Contact your school administrator.' };
		}

		const digitalId = await idRes.json();
		return { digitalId, error: null };
	} catch {
		return { digitalId: null, error: 'Could not load student data.' };
	}
};
