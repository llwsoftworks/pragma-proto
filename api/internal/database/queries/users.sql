-- users.sql: User CRUD and role queries

-- name: GetUserByEmail :one
SELECT id, school_id, role, email, password_hash, first_name, last_name,
       phone, profile_photo, mfa_secret, mfa_enabled, is_active,
       failed_login_attempts, locked_until, last_login_at, created_at, updated_at
FROM users
WHERE email = $1 AND school_id = $2
LIMIT 1;

-- name: GetUserByID :one
SELECT id, school_id, role, email, first_name, last_name,
       phone, profile_photo, mfa_enabled, is_active, last_login_at, created_at, updated_at
FROM users
WHERE id = $1 AND school_id = $2
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (school_id, role, email, password_hash, first_name, last_name, phone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, created_at;

-- name: UpdateUserMFA :exec
UPDATE users SET mfa_secret = $1, mfa_enabled = $2 WHERE id = $3 AND school_id = $4;

-- name: RecordFailedLogin :exec
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE
        WHEN failed_login_attempts + 1 >= 15 THEN NOW() + INTERVAL '100 years'
        WHEN failed_login_attempts + 1 >= 5  THEN NOW() + INTERVAL '15 minutes'
        ELSE locked_until
    END
WHERE id = $1;

-- name: ResetLoginAttempts :exec
UPDATE users
SET failed_login_attempts = 0, locked_until = NULL, last_login_at = NOW()
WHERE id = $1;

-- name: ChangePassword :exec
UPDATE users SET password_hash = $1 WHERE id = $2 AND school_id = $3;

-- name: ListUsersByRole :many
SELECT id, email, first_name, last_name, is_active, last_login_at, created_at
FROM users
WHERE school_id = $1 AND role = $2
ORDER BY last_name, first_name
LIMIT $3 OFFSET $4;

-- name: DeactivateUser :exec
UPDATE users SET is_active = FALSE WHERE id = $1 AND school_id = $2;

-- name: GetParentStudentLink :one
SELECT ps.id, ps.can_view_grades, ps.can_generate_docs, ps.relationship
FROM parent_students ps
WHERE ps.parent_id = $1 AND ps.student_id = $2;

-- name: ListParentChildren :many
SELECT s.id, u.first_name, u.last_name, s.grade_level, s.is_grade_locked,
       s.student_number, ps.can_view_grades, ps.can_generate_docs, ps.relationship
FROM parent_students ps
JOIN students s ON s.id = ps.student_id
JOIN users u ON u.id = s.user_id
WHERE ps.parent_id = $1 AND ps.school_id = $2;
