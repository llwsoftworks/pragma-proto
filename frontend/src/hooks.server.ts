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
		// Attach raw token to locals so +layout.server.ts files can forward it.
		event.locals.sessionToken = sessionCookie;

		// Decode the JWT payload (no verification — Go API verifies on every call).
		try {
			const parts = sessionCookie.split('.');
			if (parts.length === 3) {
				const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/')));
				event.locals.user = {
					id: payload.uid,
					schoolId: payload.sid,
					role: payload.role,
					email: payload.email,
					mfaDone: payload.mfa_done
				};
			}
		} catch {
			// Malformed token — treat as unauthenticated.
		}
	}

	const response = await resolve(event);

	// Security headers (spec §7.7).
	response.headers.set('X-Frame-Options', 'DENY');
	response.headers.set('X-Content-Type-Options', 'nosniff');
	response.headers.set('Referrer-Policy', 'strict-origin-when-cross-origin');
	response.headers.set('Permissions-Policy', 'camera=(), microphone=(), geolocation=()');
	response.headers.set('Strict-Transport-Security', 'max-age=31536000; includeSubDomains');
	response.headers.set(
		'Content-Security-Policy',
		"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self' https://fonts.gstatic.com"
	);

	// Request logging.
	const duration = Date.now() - start;
	console.log(`[${new Date().toISOString()}] ${event.request.method} ${event.url.pathname} ${response.status} ${duration}ms`);

	return response;
};
