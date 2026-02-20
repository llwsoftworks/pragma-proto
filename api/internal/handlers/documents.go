package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/models"
	"github.com/pragma-proto/api/internal/services"
)

// DocumentsHandler manages document generation.
type DocumentsHandler struct {
	db           *pgxpool.Pool
	pdf          *services.PDFService
	storage      *services.StorageService
	verification *services.VerificationService
	baseURL      string
}

// NewDocumentsHandler creates a DocumentsHandler.
func NewDocumentsHandler(db *pgxpool.Pool, pdf *services.PDFService, storage *services.StorageService,
	verification *services.VerificationService, baseURL string) *DocumentsHandler {
	return &DocumentsHandler{db: db, pdf: pdf, storage: storage, verification: verification, baseURL: baseURL}
}

// GenerateDocument creates an official school document and stores it in R2.
func (h *DocumentsHandler) GenerateDocument(w http.ResponseWriter, r *http.Request) {
	claims, _ := auth.ClaimsFromContext(r.Context())

	var req struct {
		StudentID string `json:"student_id" validate:"required,uuid"`
		Type      string `json:"type" validate:"required,oneof=enrollment_certificate attendance_letter academic_standing tuition_confirmation custom"`
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	ctx := r.Context()

	// Authorization: students for themselves, parents for linked children, admins for anyone.
	if claims.Role == models.RoleStudent {
		var studentUserID uuid.UUID
		h.db.QueryRow(ctx, `SELECT user_id FROM students WHERE id = $1 AND school_id = $2`,
			req.StudentID, claims.SchoolID).Scan(&studentUserID)
		if studentUserID != claims.UserID {
			writeError(w, http.StatusForbidden, "forbidden", "students can only generate documents for themselves")
			return
		}
	} else if claims.Role == models.RoleParent {
		var linkCount int
		h.db.QueryRow(ctx, `SELECT COUNT(*) FROM parent_students WHERE parent_id = $1 AND student_id = $2 AND can_generate_docs = TRUE`,
			claims.UserID, req.StudentID).Scan(&linkCount)
		if linkCount == 0 {
			writeError(w, http.StatusForbidden, "forbidden", "you are not authorized to generate documents for this student")
			return
		}
	}

	// Fetch student and school data.
	var student models.Student
	var user models.User
	var school models.School

	h.db.QueryRow(ctx, `
		SELECT s.id, s.student_number, s.grade_level, s.enrollment_status,
		       u.first_name, u.last_name, u.email
		FROM students s JOIN users u ON u.id = s.user_id
		WHERE s.id = $1 AND s.school_id = $2
	`, req.StudentID, claims.SchoolID).Scan(
		&student.ID, &student.StudentNumber, &student.GradeLevel, &student.EnrollmentStatus,
		&user.FirstName, &user.LastName, &user.Email,
	)

	h.db.QueryRow(ctx, `SELECT id, name, logo_url FROM schools WHERE id = $1`, claims.SchoolID).Scan(
		&school.ID, &school.Name, &school.LogoURL,
	)

	// Generate a document ID and verification code.
	docID := uuid.New()
	verCode := h.verification.GenerateCode(docID, claims.SchoolID)
	verURL := h.baseURL + "/verify/" + verCode

	now := time.Now()
	var expiresAt *time.Time
	if req.Type == "enrollment_certificate" {
		exp := now.AddDate(0, 6, 0) // enrollment certs expire in 6 months
		expiresAt = &exp
	}

	data := services.DocumentData{
		School:           &school,
		Student:          &student,
		StudentUser:      &user,
		DocumentType:     req.Type,
		VerificationCode: verCode,
		VerificationURL:  verURL,
		GeneratedAt:      now,
		ExpiresAt:        expiresAt,
		SignatoryName:    school.Settings.SignatoryName,
		SignatoryTitle:   school.Settings.SignatoryTitle,
	}

	htmlBytes, err := h.pdf.RenderDocumentHTML(data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "pdf_error", err.Error())
		return
	}

	// Store the HTML as PDF placeholder (in production this would be fed to chromedp).
	key := services.ObjectKey(claims.SchoolID.String(), "documents", docID.String()+".html")
	if err := h.storage.PutObject(ctx, key, htmlBytes, "text/html"); err != nil {
		writeError(w, http.StatusInternalServerError, "storage_error", err.Error())
		return
	}

	// Record the document.
	_, err = h.db.Exec(ctx, `
		INSERT INTO documents (id, school_id, student_id, type, verification_code, pdf_url, generated_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, docID, claims.SchoolID, req.StudentID, req.Type, verCode, key, claims.UserID, expiresAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}

	// Generate presigned download URL.
	downloadURL, _ := h.storage.PresignDownload(ctx, key)

	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   claims.SchoolID,
		UserID:     &claims.UserID,
		Action:     "document.generate",
		EntityType: "document",
		EntityID:   &docID,
		NewValue:   map[string]string{"type": req.Type, "student_id": req.StudentID},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"document_id":       docID,
		"verification_code": verCode,
		"download_url":      downloadURL,
		"expires_at":        expiresAt,
	})
}

// VerifyDocument is the public document verification endpoint — no auth required.
func (h *DocumentsHandler) VerifyDocument(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ctx := r.Context()

	var docType string
	var studentFirstName, studentLastName string
	var createdAt time.Time
	var expiresAt *time.Time
	var isExpired bool

	err := h.db.QueryRow(ctx, `
		SELECT d.type, u.first_name, u.last_name, d.created_at, d.expires_at
		FROM documents d
		JOIN students s ON s.id = d.student_id
		JOIN users u ON u.id = s.user_id
		WHERE d.verification_code = $1
	`, code).Scan(&docType, &studentFirstName, &studentLastName, &createdAt, &expiresAt)

	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"valid": false})
		return
	}

	if expiresAt != nil && time.Now().After(*expiresAt) {
		isExpired = true
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"valid":         !isExpired,
		"document_type": docType,
		"student_name":  studentFirstName + " " + studentLastName,
		"issued_at":     createdAt.Format("2006-01-02"),
		"expires_at":    expiresAt,
	})
}
