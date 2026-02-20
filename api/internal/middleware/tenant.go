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
func TenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := auth.ClaimsFromContext(r.Context())
		if !ok {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), tenantKey{}, claims.SchoolID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SchoolIDFromContext retrieves the tenant school_id from the request context.
func SchoolIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(tenantKey{}).(uuid.UUID)
	return id, ok
}
