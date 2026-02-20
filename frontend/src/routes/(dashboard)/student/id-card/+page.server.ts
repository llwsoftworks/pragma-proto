import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

	// Fetch the student's own digital ID.
	try {
		// We need the student's record ID first.
		const studentRes = await fetch(`${API_BASE}/students/me`, {
			headers: { Authorization: `Bearer ${locals.sessionToken}` }
		});
		if (!studentRes.ok) {
			return { digitalId: null, error: 'Could not load student data' };
		}
		const student = await studentRes.json();
		const studentId = student.id;

		const idRes = await fetch(`${API_BASE}/students/${studentId}/digital-id`, {
			headers: { Authorization: `Bearer ${locals.sessionToken}` }
		});

		if (!idRes.ok) {
			return { digitalId: null, error: 'No digital ID found. Contact your school administrator.' };
		}

		const digitalId = await idRes.json();
		return { digitalId, error: null };
	} catch {
		return { digitalId: null, error: 'Failed to load digital ID' };
	}
};
