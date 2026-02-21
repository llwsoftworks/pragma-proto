/**
 * Shared display-only utilities.
 * Grade calculations are NEVER done here — Go API handles all grade math.
 */

/** Format a date string or Date for display. */
export function formatDate(date: string | Date | null | undefined): string {
	if (!date) return '—';
	const d = typeof date === 'string' ? new Date(date) : date;
	return d.toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' });
}

/** Format a time string "HH:MM" to "h:mm AM/PM". */
export function formatTime(time: string | null | undefined): string {
	if (!time) return '';
	const [h, m] = time.split(':').map(Number);
	const period = h >= 12 ? 'PM' : 'AM';
	const hour = h % 12 || 12;
	return `${hour}:${m.toString().padStart(2, '0')} ${period}`;
}

/** Format a file size in bytes to a human-readable string. */
export function formatFileSize(bytes: number): string {
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

/** Return a CSS color class for a letter grade. */
export function gradeColor(letter: string | null | undefined): string {
	if (!letter) return 'text-muted-foreground';
	const first = letter[0].toUpperCase();
	if (first === 'A') return 'text-green-600 dark:text-green-400';
	if (first === 'B') return 'text-blue-600 dark:text-blue-400';
	if (first === 'C') return 'text-yellow-600 dark:text-yellow-400';
	if (first === 'D') return 'text-orange-600 dark:text-orange-400';
	return 'text-red-600 dark:text-red-400'; // F
}

/** Returns true if a role is in the given set. */
export function hasRole(role: string | undefined, ...roles: string[]): boolean {
	return !!role && roles.includes(role);
}

/** Format a decimal percentage for display. */
export function formatPercent(value: number | null | undefined): string {
	if (value == null) return '—';
	return `${value.toFixed(1)}%`;
}

/** Format a GPA (4.0 scale). */
export function formatGPA(gpa: number | null | undefined): string {
	if (gpa == null) return '—';
	return gpa.toFixed(3);
}

/** Day of week label. */
export const DAYS_OF_WEEK = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

/** Assignment category labels. */
export const CATEGORY_LABELS: Record<string, string> = {
	homework: 'Homework',
	quiz: 'Quiz',
	test: 'Test',
	exam: 'Exam',
	project: 'Project',
	classwork: 'Classwork',
	participation: 'Participation',
	other: 'Other'
};

/**
 * Encode a UUID to a URL-safe short ID (22 chars).
 * Strips dashes, converts hex to bytes, then base64url encodes.
 * "c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3" → "w8PDw8PDw8PDw8PDw8PDww"
 */
export function encodeId(uuid: string): string {
	const hex = uuid.replace(/-/g, '');
	const bytes = new Uint8Array(16);
	for (let i = 0; i < 16; i++) {
		bytes[i] = parseInt(hex.substring(i * 2, i * 2 + 2), 16);
	}
	let binary = '';
	for (const b of bytes) binary += String.fromCharCode(b);
	return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

/**
 * Decode a short ID back to a UUID string.
 * "w8PDw8PDw8PDw8PDw8PDww" → "c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3"
 */
export function decodeId(shortId: string): string {
	const padded = shortId.replace(/-/g, '+').replace(/_/g, '/') + '==';
	const binary = atob(padded);
	let hex = '';
	for (let i = 0; i < binary.length; i++) {
		hex += binary.charCodeAt(i).toString(16).padStart(2, '0');
	}
	return [
		hex.substring(0, 8),
		hex.substring(8, 12),
		hex.substring(12, 16),
		hex.substring(16, 20),
		hex.substring(20, 32)
	].join('-');
}

/** Clamp a number between min and max. */
export function clamp(value: number, min: number, max: number): number {
	return Math.min(Math.max(value, min), max);
}

/** Generate a debounced version of a function. */
export function debounce<T extends (...args: unknown[]) => unknown>(
	fn: T,
	delay: number
): (...args: Parameters<T>) => void {
	let timer: ReturnType<typeof setTimeout>;
	return (...args) => {
		clearTimeout(timer);
		timer = setTimeout(() => fn(...args), delay);
	};
}
