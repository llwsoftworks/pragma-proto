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
// Super-admin users have no school_id in their JWT (it will be uuid.Nil).
// They MUST provide an X-School-ID header when accessing school-scoped
// endpoints, or the middleware passes through with uuid.Nil (which the
// platform-level endpoints handle).
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		schoolID := claims.SchoolID

		// Super-admins: use X-School-ID header if provided, otherwise keep
		// uuid.Nil (acceptable for platform-level routes).
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
