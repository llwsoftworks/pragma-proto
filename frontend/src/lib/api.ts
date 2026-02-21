/**
 * Typed Go API client.
 * All state-changing operations go: Svelte UI → +page.server.ts → this client → Go API → database.
 * The frontend NEVER calls the Go API directly from the browser.
 * This file is called ONLY from +page.server.ts files (server-side).
 */

const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

export class APIError extends Error {
	constructor(
		public readonly status: number,
		public readonly code: string,
		message: string
	) {
		super(message);
	}
}

interface FetchOptions extends RequestInit {
	token?: string; // JWT for the Authorization header
}

/** Core fetch wrapper: forwards auth token, checks for errors. */
async function apiFetch<T>(path: string, options: FetchOptions = {}): Promise<T> {
	const { data } = await apiFetchRaw<T>(path, options);
	return data;
}

/**
 * Like apiFetch but also returns the raw Set-Cookie header value.
 * Used for auth endpoints (login, MFA verify) so +page.server.ts can
 * forward the session cookie to the browser via SvelteKit's cookies API.
 * Without this, the Go API's Set-Cookie is silently discarded by the
 * server-side fetch and the browser never receives the session token.
 */
async function apiFetchRaw<T>(
	path: string,
	options: FetchOptions = {}
): Promise<{ data: T; setCookie: string | null }> {
	const { token, ...init } = options;

	const headers = new Headers(init.headers);
	headers.set('Content-Type', 'application/json');
	if (token) {
		headers.set('Authorization', `Bearer ${token}`);
	}

	const response = await fetch(`${API_BASE}${path}`, { ...init, headers });

	if (!response.ok) {
		let body: { error?: string; message?: string } = {};
		try {
			body = await response.json();
		} catch {
			// Non-JSON error body.
		}
		throw new APIError(response.status, body.error ?? 'unknown_error', body.message ?? response.statusText);
	}

	const setCookie = response.headers.get('set-cookie');

	if (response.status === 204) {
		return { data: undefined as T, setCookie };
	}

	const data = (await response.json()) as T;
	return { data, setCookie };
}

// ---- Auth ----

export interface LoginResponse {
	user?: User;
	mfa_required?: boolean;
	user_id?: string;
}

export interface User {
	id: string;
	email: string;
	role: string;
	first_name: string;
	last_name: string;
	school_id: string;
}

export const auth = {
	// Returns setCookie so +page.server.ts can forward it to the browser.
	login: (email: string, password: string) =>
		apiFetchRaw<LoginResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify({ email, password })
		}),

	// Returns setCookie for the upgraded MFA-complete session token.
	verifyMFA: (code: string, token: string) =>
		apiFetchRaw<{ ok: boolean }>('/auth/mfa/verify', {
			method: 'POST',
			body: JSON.stringify({ code }),
			token
		}),

	logout: (token: string) =>
		apiFetch<void>('/auth/logout', { method: 'POST', token }),

	register: (data: RegisterData, token: string) =>
		apiFetch<{ user_id: string }>('/auth/register', {
			method: 'POST',
			body: JSON.stringify(data),
			token
		})
};

export interface RegisterData {
	school_id: string;
	role: string;
	email: string;
	password: string;
	first_name: string;
	last_name: string;
	phone?: string;
}

// ---- Dashboard ----

export const dashboard = {
	get: (token: string) => apiFetch<DashboardData>('/dashboard', { token })
};

export interface DashboardData {
	role: string;
	today_schedule?: ScheduleBlock[];
	ungraded_assignments?: number;
	recent_grade_activity?: RecentActivity[];
	children?: ChildSummary[];
	student_id?: string;
	is_grade_locked?: boolean;
	total_students?: number;
	total_teachers?: number;
	locked_students?: number;
}

// ---- Grades ----

export interface Grade {
	id: string;
	assignment_id: string;
	student_id: string;
	points_earned: number | null;
	letter_grade: string | null;
	comment: string | null;
	is_excused: boolean;
	is_missing: boolean;
	is_late: boolean;
	ai_suggested: number | null;
	ai_accepted: boolean | null;
	updated_at: string;
}

export const grades = {
	listForCourse: (courseId: string, token: string) =>
		apiFetch<{ grades: Grade[] }>(`/courses/${courseId}/grades`, { token }),

	upsert: (courseId: string, data: UpsertGradeData, token: string) =>
		apiFetch<{ grade_id: string }>(`/courses/${courseId}/grades`, {
			method: 'POST',
			body: JSON.stringify(data),
			token
		}),

	getForStudent: (studentId: string, token: string) =>
		apiFetch<{ grades: Grade[] }>(`/students/${studentId}/grades`, { token })
};

export interface UpsertGradeData {
	assignment_id: string;
	student_id: string;
	points_earned?: number | null;
	comment?: string;
	is_excused?: boolean;
	is_missing?: boolean;
	is_late?: boolean;
	ai_accepted?: boolean | null;
}

// ---- Assignments ----

export interface Assignment {
	id: string;
	short_id: string;
	course_id: string;
	title: string;
	description: string | null;
	due_date: string | null;
	max_points: number;
	category: string;
	weight: number;
	is_published: boolean;
}

export interface Attachment {
	id: string;
	file_name: string;
	file_size: number;
	mime_type: string;
	version: number;
	download_url: string;
	created_at: string;
}

export interface AssignmentListItem extends Assignment {
	course_name: string;
}

export const assignments = {
	list: (token: string) =>
		apiFetch<{ assignments: AssignmentListItem[] }>('/assignments', { token }),

	create: (data: CreateAssignmentData, token: string) =>
		apiFetch<{ assignment_id: string }>('/assignments', {
			method: 'POST',
			body: JSON.stringify(data),
			token
		}),

	requestUploadURL: (assignmentId: string, data: UploadURLRequest, token: string) =>
		apiFetch<{ upload_url: string; attachment_id: string; file_key: string }>(
			`/assignments/${assignmentId}/attachments/upload-url`,
			{ method: 'POST', body: JSON.stringify(data), token }
		),

	listAttachments: (assignmentId: string, token: string) =>
		apiFetch<{ attachments: Attachment[] }>(
			`/assignments/${assignmentId}/attachments`,
			{ token }
		)
};

export interface CreateAssignmentData {
	course_id: string;
	title: string;
	description?: string;
	due_date?: string;
	max_points: number;
	category: string;
	weight?: number;
	is_published?: boolean;
}

export interface UploadURLRequest {
	file_name: string;
	mime_type: string;
	file_size_bytes: number;
}

// ---- Admin ----

export interface StudentRow {
	id: string;
	email: string;
	first_name: string;
	last_name: string;
	student_number: string;
	grade_level: string;
	enrollment_status: string;
	is_grade_locked: boolean;
	lock_reason: string | null;
	enrollment_date: string;
}

export const admin = {
	listStudents: (token: string) =>
		apiFetch<{ students: StudentRow[] }>('/admin/students', { token }),

	lockGrade: (studentId: string, reason: string, token: string) =>
		apiFetch<{ lock_id: string }>(`/admin/students/${studentId}/lock`, {
			method: 'POST',
			body: JSON.stringify({ reason }),
			token
		}),

	unlockGrade: (studentId: string, token: string) =>
		apiFetch<{ ok: boolean }>(`/admin/students/${studentId}/lock`, {
			method: 'DELETE',
			token
		}),

	bulkLock: (studentIds: string[], reason: string, token: string) =>
		apiFetch<{ locked: number }>('/admin/grade-locks/bulk', {
			method: 'POST',
			body: JSON.stringify({ student_ids: studentIds, reason }),
			token
		})
};

// ---- Documents ----

export const documents = {
	generate: (data: GenerateDocumentData, token: string) =>
		apiFetch<{ document_id: string; verification_code: string; download_url: string }>(
			'/documents',
			{ method: 'POST', body: JSON.stringify(data), token }
		),

	verify: (code: string) =>
		apiFetch<{ valid: boolean; document_type?: string; student_name?: string; issued_at?: string }>(
			`/verify/${code}`
		)
};

export interface GenerateDocumentData {
	student_id: string;
	type: 'enrollment_certificate' | 'attendance_letter' | 'academic_standing' | 'tuition_confirmation' | 'custom';
}

// ---- AI ----

export const ai = {
	gradingAssistant: (data: GradingAssistantRequest, token: string) =>
		apiFetch<GradingAssistantResponse>('/ai/grading-assistant', {
			method: 'POST',
			body: JSON.stringify(data),
			token
		}),

	reportComment: (data: ReportCommentRequest, token: string) =>
		apiFetch<{ comment: string; ai_assisted: boolean }>('/ai/report-comment', {
			method: 'POST',
			body: JSON.stringify(data),
			token
		})
};

export interface GradingAssistantRequest {
	assignment_id: string;
	rubric: string;
	submissions: Record<string, string>;
}

export interface GradingAssistantResponse {
	raw_response: string;
	anonymized: boolean;
	student_map: Record<string, string>;
	tokens_used: number;
}

export interface ReportCommentRequest {
	student_id: string;
	course_id: string;
	grade_summary: string;
	trend_direction: 'improving' | 'declining' | 'stable';
}

// ---- Courses ----

export interface Course {
	id: string;
	short_id: string;
	name: string;
	subject: string;
	period: string | null;
	room: string | null;
	academic_year: string;
	semester: string | null;
	is_active: boolean;
	enrollment_count?: number;
}

export const courses = {
	mine: (token: string) =>
		apiFetch<{ courses: Course[] }>('/courses/mine', { token })
};

// ---- Students ----

export interface MyStudentRecord {
	id: string;
	student_number: string;
	grade_level: string;
	enrollment_status: string;
	is_grade_locked: boolean;
	enrollment_date: string;
}

export const students = {
	me: (token: string) => apiFetch<MyStudentRecord>('/students/me', { token })
};

// ---- Shared types ----

export interface ScheduleBlock {
	course_id: string | null;
	course_name: string;
	start_time: string;
	end_time: string;
	room: string | null;
	label: string | null;
	color: string | null;
}

export interface RecentActivity {
	assignment_title: string;
	course_name: string;
	average_percent: number;
	graded_count: number;
	total_count: number;
}

export interface ChildSummary {
	student_id: string;
	first_name: string;
	last_name: string;
	grade_level: string;
	is_grade_locked: boolean;
	can_view_grades: boolean;
}
