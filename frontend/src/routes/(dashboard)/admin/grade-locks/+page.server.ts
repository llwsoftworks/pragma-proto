import { fail } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { admin } from '$lib/api';

export const load: PageServerLoad = async ({ locals }) => {
	const { students } = await admin.listStudents(locals.sessionToken!);
	return { students: students ?? [] };
};

export const actions: Actions = {
	lock: async ({ request, locals }) => {
		const data = await request.formData();
		const studentId = data.get('student_id')?.toString() ?? '';
		const reason = data.get('reason')?.toString() ?? '';

		if (!studentId || !reason) {
			return fail(400, { error: 'Student and reason are required' });
		}

		try {
			await admin.lockGrade(studentId, reason, locals.sessionToken!);
			return { success: true, action: 'locked' };
		} catch (err: unknown) {
			return fail(500, { error: (err as Error).message });
		}
	},

	unlock: async ({ request, locals }) => {
		const data = await request.formData();
		const studentId = data.get('student_id')?.toString() ?? '';

		try {
			await admin.unlockGrade(studentId, locals.sessionToken!);
			return { success: true, action: 'unlocked' };
		} catch (err: unknown) {
			return fail(500, { error: (err as Error).message });
		}
	}
};
