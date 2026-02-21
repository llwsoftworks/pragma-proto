import { error } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';

const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

export const load: PageServerLoad = async ({ params, locals }) => {
	// params.courseId is already the 8-char short_id — pass it directly to the API.
	const courseId = params.courseId;
	const token = locals.sessionToken!;

	try {
		// Fetch course details and enrolled students in parallel.
		const [courseRes, studentsRes, assignmentsRes, gradesRes] = await Promise.all([
			fetch(`${API_BASE}/courses/${courseId}`, {
				headers: { Authorization: `Bearer ${token}` }
			}),
			fetch(`${API_BASE}/courses/${courseId}/students`, {
				headers: { Authorization: `Bearer ${token}` }
			}),
			fetch(`${API_BASE}/courses/${courseId}/assignments`, {
				headers: { Authorization: `Bearer ${token}` }
			}),
			fetch(`${API_BASE}/courses/${courseId}/grades`, {
				headers: { Authorization: `Bearer ${token}` }
			})
		]);

		const course = courseRes.ok ? await courseRes.json() : null;
		if (!course) {
			throw error(404, 'Course not found');
		}

		const studentsData = studentsRes.ok ? await studentsRes.json() : { students: [] };
		const assignmentsData = assignmentsRes.ok ? await assignmentsRes.json() : { assignments: [] };
		const gradesData = gradesRes.ok ? await gradesRes.json() : { grades: [] };

		return {
			course,
			students: studentsData.students ?? [],
			assignments: assignmentsData.assignments ?? [],
			grades: gradesData.grades ?? []
		};
	} catch (err) {
		if (err && typeof err === 'object' && 'status' in err) throw err;
		return {
			course: { id: '', short_id: courseId, name: 'Unknown Course', subject: '' },
			students: [],
			assignments: [],
			grades: []
		};
	}
};

export const actions: Actions = {
	saveGrade: async ({ request, locals, params }) => {
		const data = await request.formData();
		const assignmentId = data.get('assignment_id')?.toString() ?? '';
		const studentId = data.get('student_id')?.toString() ?? '';
		const pointsRaw = data.get('points_earned')?.toString() ?? '';
		const isExcused = data.get('is_excused') === 'true';
		const isMissing = data.get('is_missing') === 'true';
		const isLate = data.get('is_late') === 'true';

		const points = pointsRaw === '' ? null : Number(pointsRaw);
		const courseId = params.courseId; // short_id — passed directly to API

		try {
			const token = locals.sessionToken!;
			await fetch(`${API_BASE}/courses/${courseId}/grades`, {
				method: 'POST',
				headers: {
					Authorization: `Bearer ${token}`,
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					assignment_id: assignmentId,
					student_id: studentId,
					points_earned: points,
					is_excused: isExcused,
					is_missing: isMissing,
					is_late: isLate
				})
			});
			return { success: true };
		} catch {
			return { success: false, error: 'Failed to save grade' };
		}
	}
};
