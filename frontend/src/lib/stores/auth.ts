import { writable } from 'svelte/store';

export interface AuthUser {
	id: string;
	email: string;
	role: string;
	schoolId: string;
	firstName?: string;
	lastName?: string;
	mfaDone: boolean;
}

/** Current authenticated user. Set by the dashboard layout after SSR. */
export const currentUser = writable<AuthUser | null>(null);

/** Clear the user session (call on logout). */
export function clearUser() {
	currentUser.set(null);
}
