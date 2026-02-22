package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pragma-proto/api/internal/auth"
	"github.com/pragma-proto/api/internal/middleware"
	"github.com/pragma-proto/api/internal/models"
)

var validate = validator.New()

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSvc    *auth.JWTService
	encryptor *auth.LoginEncryptor
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(db *pgxpool.Pool, jwtSvc *auth.JWTService, encryptor *auth.LoginEncryptor) *AuthHandler {
	return &AuthHandler{db: db, jwtSvc: jwtSvc, encryptor: encryptor}
}

// loginRequest is validated strictly — unknown fields are rejected.
type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=1"`
}

// encryptedLoginRequest wraps the AES-256-GCM encrypted login payload.
type encryptedLoginRequest struct {
	Encrypted string `json:"encrypted" validate:"required"`
}

// Login authenticates a user and sets the session cookie.
// Accepts an AES-256-GCM encrypted payload: {"encrypted": "<base64>"}
// The encrypted payload decrypts to the standard loginRequest JSON.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var encReq encryptedLoginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&encReq); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Decrypt the payload.
	plaintext, err := h.encryptor.Decrypt(encReq.Encrypted)
	if err != nil {
		writeError(w, http.StatusBadRequest, "decryption_failed", "unable to decrypt login payload")
		return
	}

	var req loginRequest
	if err := json.Unmarshal(plaintext, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "decrypted payload is not valid JSON")
		return
	}
	if err := validate.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, "validation_error", err.Error())
		return
	}

	ctx := r.Context()

	// Look up the user by email across all schools (email alone is not unique globally,
	// so we require the user to also provide their school identifier in a real flow).
	// For simplicity, this prototype looks up by email only — in production the login page
	// would include a school selector that resolves to school_id.
	var user models.User
	err := h.db.QueryRow(ctx, `
		SELECT id, school_id, role, email, password_hash, first_name, last_name,
		       mfa_enabled, is_active, failed_login_attempts, locked_until
		FROM users WHERE email = $1 LIMIT 1
	`, req.Email).Scan(
		&user.ID, &user.SchoolID, &user.Role, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.MFAEnabled, &user.IsActive,
		&user.FailedLoginAttempts, &user.LockedUntil,
	)
	if err != nil {
		// Use the same error message for not found and bad password (prevent user enumeration).
		auditFailedLogin(ctx, h.db, req.Email, r)
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
		return
	}

	if !user.IsActive {
		writeError(w, http.StatusForbidden, "account_inactive", "account has been deactivated")
		return
	}

	// Check account lockout.
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		writeError(w, http.StatusTooManyRequests, "account_locked", "account is temporarily locked due to too many failed attempts")
		return
	}

	ok, err := auth.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil || !ok {
		h.recordFailedLogin(ctx, user.ID, user.FailedLoginAttempts)
		auditFailedLogin(ctx, h.db, req.Email, r)
		writeError(w, http.StatusUnauthorized, "invalid_credentials", "email or password is incorrect")
		return
	}

	// Reset failed attempt counter on success.
	h.db.Exec(ctx, `UPDATE users SET failed_login_attempts = 0, locked_until = NULL, last_login_at = NOW() WHERE id = $1`, user.ID)

	// For MFA-required roles, issue a partial token (mfa_done=false).
	mfaDone := !user.MFAEnabled
	if user.Role == models.RoleParent || user.Role == models.RoleStudent {
		mfaDone = true // MFA optional for these roles
	}

	// Resolve school_id: super_admins have NULL, use zero UUID in JWT.
	schoolID := uuid.Nil
	if user.SchoolID != nil {
		schoolID = *user.SchoolID
	}

	token, err := h.jwtSvc.Issue(user.ID, schoolID, user.Role, user.Email, mfaDone)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_error", "failed to issue token")
		return
	}

	// Store session hash.
	h.storeSession(ctx, user.ID, user.SchoolID, token, r)

	setSessionCookie(w, token, user.Role)

	if !mfaDone {
		// Redirect to MFA verification.
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"mfa_required": true,
			"user_id":      user.ID,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         user.ID,
			"email":      user.Email,
			"role":       user.Role,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"school_id":  user.SchoolID,
		},
	})
}

// VerifyMFA validates a TOTP code and upgrades the session token to mfa_done=true.
func (h *AuthHandler) VerifyMFA(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized", "")
		return
	}

	var req struct {
		Code string `json:"code" validate:"required,len=6"`
	}
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	ctx := r.Context()

	var mfaSecret string
	err := h.db.QueryRow(ctx, `SELECT mfa_secret FROM users WHERE id = $1`, claims.UserID).Scan(&mfaSecret)
	if err != nil || mfaSecret == "" {
		writeError(w, http.StatusBadRequest, "mfa_not_setup", "MFA is not configured")
		return
	}

	if !auth.VerifyTOTP(mfaSecret, req.Code) {
		writeError(w, http.StatusUnauthorized, "invalid_mfa_code", "the MFA code is incorrect or expired")
		return
	}

	// Issue a new token with mfa_done=true.
	token, err := h.jwtSvc.Issue(claims.UserID, claims.SchoolID, claims.Role, claims.Email, true)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token_error", "")
		return
	}

	// SchoolID may be nil for super_admins.
	var schoolIDPtr *uuid.UUID
	if claims.SchoolID != uuid.Nil {
		schoolIDPtr = &claims.SchoolID
	}
	h.storeSession(ctx, claims.UserID, schoolIDPtr, token, r)
	setSessionCookie(w, token, claims.Role)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// Logout invalidates the session.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		http.SetCookie(w, &http.Cookie{Name: "session", MaxAge: -1})
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Invalidate all sessions for this user.
	h.db.Exec(r.Context(), `DELETE FROM sessions WHERE user_id = $1`, claims.UserID)

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusNoContent)
}

// Register creates a new user account (super_admin only in production — or during onboarding).
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		SchoolID  string `json:"school_id" validate:"required,uuid"`
		Role      string `json:"role" validate:"required,oneof=admin teacher parent student"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required,min=12"`
		FirstName string `json:"first_name" validate:"required,min=1,max=100"`
		LastName  string `json:"last_name" validate:"required,min=1,max=100"`
		Phone     string `json:"phone"`
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

	if err := auth.ValidatePasswordStrength(req.Password); err != nil {
		writeError(w, http.StatusBadRequest, "weak_password", err.Error())
		return
	}

	breached, _ := auth.CheckBreachedPassword(req.Password)
	if breached {
		writeError(w, http.StatusBadRequest, "breached_password",
			"this password has appeared in a known data breach; please choose a different password")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "hash_error", "")
		return
	}

	ctx := r.Context()
	var userID uuid.UUID
	err = h.db.QueryRow(ctx, `
		INSERT INTO users (school_id, role, email, password_hash, first_name, last_name, phone)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.SchoolID, req.Role, req.Email, hash, req.FirstName, req.LastName, nullStr(req.Phone)).Scan(&userID)
	if err != nil {
		writeError(w, http.StatusConflict, "email_exists", "email already registered at this school")
		return
	}

	// Audit log.
	schoolUUID, _ := uuid.Parse(req.SchoolID)
	_ = middleware.WriteAuditLog(ctx, h.db, middleware.AuditEntry{
		SchoolID:   schoolUUID,
		UserID:     &userID,
		Action:     "user.register",
		EntityType: "user",
		EntityID:   &userID,
		NewValue:   map[string]string{"email": req.Email, "role": req.Role},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, map[string]interface{}{"user_id": userID})
}

// ---- helpers ----

func (h *AuthHandler) recordFailedLogin(ctx context.Context, userID uuid.UUID, attempts int) {
	next := attempts + 1
	var lockedUntil interface{}
	if next >= 15 {
		// Admin reset required — lock indefinitely (represented by far future timestamp).
		future := time.Now().Add(10 * 365 * 24 * time.Hour)
		lockedUntil = future
	} else if next >= 5 {
		// 15-minute lockout.
		future := time.Now().Add(15 * time.Minute)
		lockedUntil = future
	}
	h.db.Exec(ctx, `UPDATE users SET failed_login_attempts = $1, locked_until = $2 WHERE id = $3`,
		next, lockedUntil, userID)
}

func (h *AuthHandler) storeSession(ctx context.Context, userID uuid.UUID, schoolID *uuid.UUID, token string, r *http.Request) {
	hash := auth.HashToken(token)
	expiry := time.Now().Add(24 * time.Hour)
	h.db.Exec(ctx, `
		INSERT INTO sessions (user_id, school_id, token_hash, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, userID, schoolID, hash, r.RemoteAddr, r.UserAgent(), expiry)
}

func auditFailedLogin(ctx context.Context, db *pgxpool.Pool, email string, r *http.Request) {
	// We don't know the school_id on failure so we use a zero UUID for the log.
	var nilUUID uuid.UUID
	_ = middleware.WriteAuditLog(ctx, db, middleware.AuditEntry{
		SchoolID:   nilUUID,
		Action:     "user.login_failed",
		EntityType: "user",
		NewValue:   map[string]string{"email": email},
		IPAddress:  r.RemoteAddr,
		UserAgent:  r.UserAgent(),
	})
}

func setSessionCookie(w http.ResponseWriter, token, role string) {
	dur := 24 * time.Hour
	if role == models.RoleTeacher || role == models.RoleAdmin || role == models.RoleSuperAdmin {
		dur = 15 * time.Minute
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(dur.Seconds()),
	})
}

func nullStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
