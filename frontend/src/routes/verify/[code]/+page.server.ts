import type { PageServerLoad } from './$types';

/**
 * Public document/ID verification endpoint — no authentication required.
 * Displays only valid/invalid status, student name, and document type.
 * No sensitive data is exposed.
 */
export const load: PageServerLoad = async ({ params }) => {
	const code = params.code;
	const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

	try {
		const res = await fetch(`${API_BASE}/verify/${code}`);
		const data = res.ok ? await res.json() : { valid: false };
		return { verification: data, code };
	} catch {
		return { verification: { valid: false }, code };
	}
};
