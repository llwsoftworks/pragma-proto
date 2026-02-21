import type { Handle } from '@sveltejs/kit';

/**
 * Global server hook:
 * - Sets security headers on every response (spec §7.7)
 * - Parses the session JWT cookie and attaches the raw token to locals
 * - Logs every request (method, path, timing)
 */
export const handle: Handle = async ({ event, resolve }) => {
	const start = Date.now();

	// Parse session cookie.
	const sessionCookie = event.cookies.get('session');
	if (sessionCookie) {
		// Decode the JWT payload (no verification — Go API verifies on every call).
		try {
			const parts = sessionCookie.split('.');
			if (parts.length === 3) {
				const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));

				// Check expiration: if the token has expired, treat as unauthenticated
				// so the layout auth guard redirects to /login instead of letting
				// page loaders call the Go API with a stale token (which returns 401
				// and surfaces as a 500 to the user).
				const now = Math.floor(Date.now() / 1000);
				if (payload.exp && payload.exp < now) {
					event.cookies.delete('session', { path: '/' });
				} else {
					event.locals.sessionToken = sessionCookie;
					event.locals.user = {
						id: payload.uid,
						schoolId: payload.sid,
						role: payload.role,
						email: payload.email,
						mfaDone: payload.mfa_done
					};
				}
			}
		} catch {
			// Malformed token — treat as unauthenticated.
		}
	}

	const response = await resolve(event);

	// Security headers (spec §7.7).
	// NOTE: Content-Security-Policy is configured in svelte.config.js via kit.csp
	// so SvelteKit can inject nonces for its inline hydration scripts. Setting
	// script-src 'self' here WITHOUT nonces blocks hydration and kills all
	// client-side interactivity.
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
	response.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');

	// Request logging.
	const duration = Date.now() - start;
	console.log(`[${new Date().toISOString()}] ${event.request.method} ${event.url.pathname} ${response.status} ${duration}ms`);

	return response;
};
