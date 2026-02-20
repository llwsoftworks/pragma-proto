package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	contextKeyClaims contextKey = "claims"
)

// Middleware validates the JWT from the HTTP-only cookie and injects Claims
// into the request context. Returns 401 if missing or invalid.
func Middleware(jwtSvc *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwtSvc.Validate(token)
			if err != nil {
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			// For MFA-required roles, block access if MFA is not yet verified.
			mfaRequired := map[string]bool{
				"super_admin": true,
				"admin":       true,
				"teacher":     true,
			}
			if mfaRequired[claims.Role] && !claims.MFADone {
				http.Error(w, `{"error":"mfa_required"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), contextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext retrieves the validated JWT Claims from the request context.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(contextKeyClaims).(*Claims)
	return c, ok
}

// extractToken pulls the JWT from the session cookie or Authorization header.
func extractToken(r *http.Request) string {
	// Prefer the HTTP-only cookie (primary auth mechanism).
	if cookie, err := r.Cookie("session"); err == nil {
		return cookie.Value
	}

	// Fallback: Authorization: Bearer <token> (for API clients).
	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}
