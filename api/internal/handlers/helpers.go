package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		// Log only — the response header is already sent.
		_ = err
	}
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]string{
		"error":   code,
		"message": message,
	})
}

// paginate extracts limit/offset from query parameters with sane defaults.
func paginate(r *http.Request) (limit, offset int) {
	limit = 50
	offset = 0

	if l := r.URL.Query().Get("limit"); l != "" {
		var lv int
		if _, err := parseIntParam(l, &lv); err == nil && lv > 0 && lv <= 200 {
			limit = lv
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		var ov int
		if _, err := parseIntParam(o, &ov); err == nil && ov >= 0 {
			offset = ov
		}
	}
	return
}

func parseIntParam(s string, out *int) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	if err != nil {
		return 0, err
	}
	*out = n
	return n, nil
}

// resolveStudentUUID looks up a student's UUID from its short_id, scoped to a school.
func resolveStudentUUID(ctx context.Context, db *pgxpool.Pool, shortID string, schoolID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM students WHERE short_id = $1 AND school_id = $2`,
		shortID, schoolID,
	).Scan(&id)
	return id, err
}

// resolveScheduleBlockUUID looks up a schedule block's UUID from its short_id.
func resolveScheduleBlockUUID(ctx context.Context, db *pgxpool.Pool, shortID string, userID, schoolID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM schedule_blocks WHERE short_id = $1 AND user_id = $2 AND school_id = $3`,
		shortID, userID, schoolID,
	).Scan(&id)
	return id, err
}

// resolveDigitalIDUUID looks up a digital ID's UUID from its short_id.
func resolveDigitalIDUUID(ctx context.Context, db *pgxpool.Pool, shortID string, schoolID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := db.QueryRow(ctx,
		`SELECT id FROM digital_ids WHERE short_id = $1 AND school_id = $2`,
		shortID, schoolID,
	).Scan(&id)
	return id, err
}
