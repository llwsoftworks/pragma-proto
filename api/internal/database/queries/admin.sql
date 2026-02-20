-- admin.sql: Grade lock and admin bulk operations

-- name: LockStudentGrades :one
INSERT INTO grade_locks (student_id, school_id, locked_by, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, locked_at;

-- name: UnlockStudentGrades :one
UPDATE grade_locks
SET is_active = FALSE, unlocked_at = NOW(), unlocked_by = $1
WHERE student_id = $2 AND school_id = $3 AND is_active = TRUE
RETURNING id;

-- name: UpdateStudentGradeLockState :exec
UPDATE students
SET is_grade_locked = $1, lock_reason = $2
WHERE id = $3 AND school_id = $4;

-- name: GetActiveLocks :many
SELECT gl.id, gl.student_id, u.first_name, u.last_name, s.student_number,
       gl.reason, gl.locked_at, gl.locked_by
FROM grade_locks gl
JOIN students s ON s.id = gl.student_id
JOIN users u ON u.id = s.user_id
WHERE gl.school_id = $1 AND gl.is_active = TRUE
ORDER BY gl.locked_at DESC;

-- name: UpdateSchoolSettings :exec
UPDATE schools SET settings = $1, updated_at = NOW() WHERE id = $2;

-- name: GetSchoolSettings :one
SELECT id, name, address, logo_url, settings FROM schools WHERE id = $1;
