// See https://kit.svelte.dev/docs/types#app
// for information about these interfaces
declare global {
	namespace App {
		interface Locals {
			sessionToken?: string;
			user?: {
				id: string;
				email: string;
				role: string;
				schoolId: string;
				mfaDone: boolean;
			};
		}
		interface PageData {}
		interface PageState {}
		interface Error {
			message: string;
			code?: string;
		}
		interface Platform {}
	}
}

export {};
