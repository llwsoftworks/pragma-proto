package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/pragma-proto/api/internal/auth"
)

type tenantKey struct{}

// TenantMiddleware extracts school_id from JWT claims and injects it into
// the request context. Every database query MUST use this value for scoping.
//
// Super-admin override: if the user's role is super_admin and the request
// includes an X-School-ID header, that value is used instead of the JWT's
// school_id. This allows super-admins to operate on any school through
// the standard admin endpoints without needing a separate JWT.
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		schoolID := claims.SchoolID

		// Allow super_admin to override school context.
		if claims.Role == "super_admin" {
			if override := r.Header.Get("X-School-ID"); override != "" {
				parsed, err := uuid.Parse(override)
				if err != nil {
					http.Error(w, `{"error":"invalid_school_id","message":"X-School-ID header must be a valid UUID"}`, http.StatusBadRequest)
					return
				}
				schoolID = parsed
			}
		}

		ctx := context.WithValue(r.Context(), tenantKey{}, schoolID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SchoolIDFromContext retrieves the tenant school_id from the request context.
func SchoolIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantKey{}).(uuid.UUID)
	return id, ok
}
