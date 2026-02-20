package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	skip "github.com/skip2/go-qrcode"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/services"
)

// DigitalIDHandler manages digital student ID cards.
type DigitalIDHandler struct {
	db           *pgxpool.Pool
	storage      *services.StorageService
	verification *services.VerificationService
	baseURL      string
}

// NewDigitalIDHandler creates a DigitalIDHandler.
func NewDigitalIDHandler(db *pgxpool.Pool, storage *services.StorageService,
	verification *services.VerificationService, baseURL string) *DigitalIDHandler {
	return &DigitalIDHandler{
		db:           db,
		storage:      storage,
		verification: verification,
		baseURL:      baseURL,
	}
}

// GetStudentID returns the digital ID card for a student.
func (h *DigitalIDHandler) GetStudentID(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentIDStr := chi.URLParam(r, "studentId")
	ctx := r.Context()

	// Students can only view their own ID.
	if claims.Role == "student" {
		var studentUserID uuid.UUID
		h.db.QueryRow(ctx, `SELECT user_id FROM students WHERE id = $1 AND school_id = $2`,
			studentIDStr, claims.SchoolID).Scan(&studentUserID)
		if studentUserID != claims.UserID {
			writeError(w, http.StatusForbidden, "forbidden", "you can only view your own ID")
			return
		}
	}

	var id struct {
		ID          uuid.UUID  `json:"id"`
		StudentID   uuid.UUID  `json:"student_id"`
		IDNumber    string     `json:"id_number"`
		QRCodeData  string     `json:"qr_code_data"`
		BarcodeData *string    `json:"barcode_data"`
		PhotoURL    *string    `json:"photo_url"`
		IssuedAt    time.Time  `json:"issued_at"`
		ExpiresAt   time.Time  `json:"expires_at"`
		IsValid     bool       `json:"is_valid"`
		FirstName   string     `json:"first_name"`
		LastName    string     `json:"last_name"`
		GradeLevel  string     `json:"grade_level"`
		SchoolName  string     `json:"school_name"`
		SchoolLogo  *string    `json:"school_logo_url"`
	}

	err := h.db.QueryRow(ctx, `
		SELECT di.id, di.student_id, di.id_number, di.qr_code_data, di.barcode_data,
		       di.photo_url, di.issued_at, di.expires_at, di.is_valid,
		       u.first_name, u.last_name, s.grade_level, sch.name, sch.logo_url
		FROM digital_ids di
		JOIN students s ON s.id = di.student_id
		JOIN users u ON u.id = s.user_id
		JOIN schools sch ON sch.id = di.school_id
		WHERE di.student_id = $1 AND di.school_id = $2 AND di.is_valid = TRUE
		ORDER BY di.issued_at DESC
		LIMIT 1
	`, studentIDStr, claims.SchoolID).Scan(
		&id.ID, &id.StudentID, &id.IDNumber, &id.QRCodeData, &id.BarcodeData,
		&id.PhotoURL, &id.IssuedAt, &id.ExpiresAt, &id.IsValid,
		&id.FirstName, &id.LastName, &id.GradeLevel, &id.SchoolName, &id.SchoolLogo,
	)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "no valid digital ID found")
		return
	}

	writeJSON(w, http.StatusOK, id)
}

// IssueStudentID generates a new digital ID for a student.
func (h *DigitalIDHandler) IssueStudentID(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	studentIDStr := chi.URLParam(r, "studentId")
	ctx := r.Context()

	studentID, err := uuid.Parse(studentIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid_student_id", "")
		return
	}

	// Generate a unique ID number.
	idNumber := generateIDNumber(claims.SchoolID, studentID)

	// Invalidate any existing active IDs.
	h.db.Exec(ctx, `UPDATE digital_ids SET is_valid = FALSE WHERE student_id = $1 AND school_id = $2`,
		studentID, claims.SchoolID)

	// Create the new ID record.
	var idID uuid.UUID
	idID = uuid.New()
	verificationURL := h.baseURL + "/verify/" + h.verification.GenerateCode(idID, claims.SchoolID)
	qrData := verificationURL

	expiresAt := time.Now().AddDate(1, 0, 0) // expires in 1 year

	var newIDID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO digital_ids (student_id, school_id, id_number, qr_code_data, barcode_data, issued_at, expires_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), $6)
		RETURNING id
	`, studentID, claims.SchoolID, idNumber, qrData, idNumber, expiresAt).Scan(&newIDID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Generate QR code image and store in R2.
	qrPNG, qrErr := skip.Encode(qrData, skip.Medium, 256)
	if qrErr == nil {
		key := services.ObjectKey(claims.SchoolID.String(), "ids", newIDID.String()+"-qr.png")
		h.storage.PutObject(ctx, key, qrPNG, "image/png")
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "digital_id.issue",
		EntityType: "digital_id",
		EntityID:   &newIDID,
		NewValue:   map[string]interface{}{"student_id": studentID, "id_number": idNumber},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})
	_ = idID

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":               newIDID,
		"id_number":        idNumber,
		"verification_url": verificationURL,
		"expires_at":       expiresAt,
	})
}

// RevokeStudentID revokes a digital ID (admin only).
func (h *DigitalIDHandler) RevokeStudentID(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())
	idIDStr := chi.URLParam(r, "idId")
	ctx := r.Context()

	var idID uuid.UUID
	err := h.db.QueryRow(ctx, `
		UPDATE digital_ids SET is_valid = FALSE, revoked_at = NOW()
		WHERE id = $1 AND school_id = $2 AND is_valid = TRUE
		RETURNING id
	`, idIDStr, claims.SchoolID).Scan(&idID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "ID not found or already revoked")
		return
	}

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "digital_id.revoke",
		EntityType: "digital_id",
		EntityID:   &idID,
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// VerifyID is the public verification endpoint — no auth required.
func (h *DigitalIDHandler) VerifyID(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ctx := r.Context()

	// Look up by QR code data (which encodes the verification URL with the code).
	var isValid bool
	var firstName, lastName string
	var photoURL *string
	var expiresAt time.Time

	err := h.db.QueryRow(ctx, `
		SELECT di.is_valid, u.first_name, u.last_name, di.photo_url, di.expires_at
		FROM digital_ids di
		JOIN students s ON s.id = di.student_id
		JOIN users u ON u.id = s.user_id
		WHERE di.qr_code_data LIKE '%' || $1 AND di.is_valid = TRUE
		LIMIT 1
	`, code).Scan(&isValid, &firstName, &lastName, &photoURL, &expiresAt)

	if err != nil || !isValid || time.Now().After(expiresAt) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"valid": false,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":        true,
		"student_name": firstName + " " + lastName,
		"photo_url":    photoURL,
		"expires_at":   expiresAt.Format("2006-01-02"),
	})
}

func generateIDNumber(schoolID, studentID uuid.UUID) string {
	// Simple deterministic ID number: school prefix + student UUID segment.
	prefix := schoolID.String()[:4]
	suffix := studentID.String()[:8]
	return prefix + "-" + suffix
}
