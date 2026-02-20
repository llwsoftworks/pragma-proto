import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { assignments } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	// Load teacher's courses for the course selector.
	const API_BASE = process.env.API_URL ?? 'http://localhost:8080';
	try {
		const res = await fetch(`${API_BASE}/courses/mine`, {
			headers: { Authorization: `Bearer ${locals.sessionToken}` }
		});
		const data = res.ok ? await res.json() : { courses: [] };
		return { courses: data.courses ?? [] };
	} catch {
		return { courses: [] };
	}
};

export const actions: Actions = {
	default: async ({ request, locals }) => {
		const data = await request.formData();

		const courseId = data.get('course_id')?.toString() ?? '';
		const title = data.get('title')?.toString() ?? '';
		const description = data.get('description')?.toString();
		const dueDate = data.get('due_date')?.toString();
		const maxPoints = parseFloat(data.get('max_points')?.toString() ?? '0');
		const category = data.get('category')?.toString() ?? 'other';
		const weight = parseFloat(data.get('weight')?.toString() ?? '1.0');
		const isPublished = data.get('is_published') === 'on';

		if (!courseId || !title || maxPoints <= 0) {
			return fail(400, { error: 'Course, title, and max points are required' });
		}

		try {
			const result = await assignments.create(
				{
					course_id: courseId,
					title,
					description: description || undefined,
					due_date: dueDate ? new Date(dueDate).toISOString() : undefined,
					max_points: maxPoints,
					category,
					weight,
					is_published: isPublished
				},
				locals.sessionToken!
			);
			throw redirect(302, `/teacher/assignments/${result.assignment_id}`);
		} catch (err: unknown) {
			if ((err as { status?: number })?.status === 302) throw err;
			return fail(500, { error: (err as Error).message ?? 'Failed to create assignment' });
		}
	}
};
