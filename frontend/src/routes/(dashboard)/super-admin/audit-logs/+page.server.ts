import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { platform } from '$lib/api';

export const load: PageServerLoad = async ({ locals, url }) => {
	if (locals.user?.role !== 'super_admin') {
		throw redirect(302, '/');
	}

	const schoolId = url.searchParams.get('school_id') ?? undefined;
	const action = url.searchParams.get('action') ?? undefined;

	const [logsData, schoolsData] = await Promise.all([
		platform.listAuditLogs(locals.sessionToken!, { school_id: schoolId, action }),
		platform.listSchools(locals.sessionToken!)
	]);

	return {
		auditLogs: logsData.audit_logs,
		schools: schoolsData.schools,
		filters: { school_id: schoolId, action }
	};
};
