package middleware

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
)

// AuditEntry holds data for an audit log row.
type AuditEntry struct {
	SchoolID   uuid.UUID
	UserID     *uuid.UUID
	Action     string
	EntityType string
	EntityID   *uuid.UUID
	OldValue   interface{}
	NewValue   interface{}
	IPAddress  string
	UserAgent  string
}

// WriteAuditLog inserts an append-only audit log entry.
func WriteAuditLog(ctx context.Context, pool *pgxpool.Pool, entry AuditEntry) error {
	var oldJSON, newJSON []byte
	var err error
	if entry.OldValue != nil {
		oldJSON, err = json.Marshal(entry.OldValue)
		if err != nil {
			return err
		}
	}
	if entry.NewValue != nil {
		newJSON, err = json.Marshal(entry.NewValue)
		if err != nil {
			return err
		}
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO audit_logs
			(school_id, user_id, action, entity_type, entity_id, old_value, new_value, ip_address, user_agent, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`,
		entry.SchoolID,
		entry.UserID,
		entry.Action,
		entry.EntityType,
		entry.EntityID,
		oldJSON,
		newJSON,
		entry.IPAddress,
		entry.UserAgent,
		time.Now(),
	)
	return err
}

// AuditMiddleware automatically logs all non-GET requests to the audit log.
// Handlers that need richer old/new value capture should call WriteAuditLog directly.
func AuditMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)

			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				return
			}

			claims, ok := auth.ClaimsFromContext(r.Context())
			if !ok {
				return
			}

			uid := claims.UserID
			entry := AuditEntry{
				SchoolID:   claims.SchoolID,
				UserID:     &uid,
				Action:     strings.ToLower(r.Method) + "." + r.URL.Path,
				EntityType: "http_request",
				IPAddress:  extractIP(r),
				UserAgent:  r.UserAgent(),
			}
			// Best-effort: ignore errors in audit middleware so they never affect responses.
			_ = WriteAuditLog(context.Background(), pool, entry)
		})
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
