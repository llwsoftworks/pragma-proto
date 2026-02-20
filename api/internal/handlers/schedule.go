package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/models"
)

// ScheduleHandler manages schedule blocks.
type ScheduleHandler struct {
	db *pgxpool.Pool
}

// NewScheduleHandler creates a ScheduleHandler.
func NewScheduleHandler(db *pgxpool.Pool) *ScheduleHandler {
	return &ScheduleHandler{db: db}
}

// ListSchedule returns schedule blocks for the current user.
func (h *ScheduleHandler) ListSchedule(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	ctx := r.Context()

	rows, err := h.db.Query(ctx, `
		SELECT sb.id, sb.course_id, COALESCE(c.name, '') AS course_name,
		       sb.day_of_week, sb.start_time::text, sb.end_time::text,
		       sb.room, sb.label, sb.color, sb.is_recurring, sb.semester
		FROM schedule_blocks sb
		LEFT JOIN courses c ON c.id = sb.course_id
		WHERE sb.user_id = $1 AND sb.school_id = $2
		ORDER BY sb.day_of_week, sb.start_time
	`, claims.UserID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	defer rows.Close()

	var blocks []models.ScheduleBlock
	for rows.Next() {
		var b models.ScheduleBlock
		if err := rows.Scan(
			&b.ID, &b.CourseID, &b.CourseName,
			&b.DayOfWeek, &b.StartTime, &b.EndTime,
			&b.Room, &b.Label, &b.Color, &b.IsRecurring, &b.Semester,
		); err != nil {
			continue
		}
		blocks = append(blocks, b)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"blocks": blocks})
}

// CreateScheduleBlock adds a new schedule block with conflict detection.
func (h *ScheduleHandler) CreateScheduleBlock(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		CourseID    *string `json:"course_id"`
		DayOfWeek   int     `json:"day_of_week" validate:"min=0,max=6"`
		StartTime   string  `json:"start_time" validate:"required"`
		EndTime     string  `json:"end_time" validate:"required"`
		Room        string  `json:"room"`
		Label       string  `json:"label"`
		Color       string  `json:"color"`
		Semester    string  `json:"semester"`
		IsRecurring bool    `json:"is_recurring"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := r.Context()

	// Check for room conflicts.
	if req.Room != "" {
		var conflictCount int
		h.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM schedule_blocks
			WHERE school_id = $1 AND room = $2 AND day_of_week = $3
			  AND start_time < $5::time AND end_time > $4::time
		`, claims.SchoolID, req.Room, req.DayOfWeek, req.StartTime, req.EndTime).Scan(&conflictCount)

		if conflictCount > 0 {
			writeError(w, http.StatusConflict, "room_conflict",
				"another block is scheduled in this room at an overlapping time")
			return
		}
	}

	var blockID uuid.UUID
	err := h.db.QueryRow(ctx, `
		INSERT INTO schedule_blocks
			(school_id, user_id, course_id, day_of_week, start_time, end_time,
			 room, label, color, semester, is_recurring)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`, claims.SchoolID, claims.UserID, req.CourseID,
		req.DayOfWeek, req.StartTime, req.EndTime,
		nullStr(req.Room), nullStr(req.Label), nullStr(req.Color),
		nullStr(req.Semester), req.IsRecurring,
	).Scan(&blockID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{"block_id": blockID})
}

// DeleteScheduleBlock removes a schedule block.
func (h *ScheduleHandler) DeleteScheduleBlock(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	blockID := chi.URLParam(r, "blockId")
	ctx := r.Context()

	_, err := h.db.Exec(ctx, `
		DELETE FROM schedule_blocks WHERE id = $1 AND user_id = $2 AND school_id = $3
	`, blockID, claims.UserID, claims.SchoolID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "block not found or not owned by you")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
