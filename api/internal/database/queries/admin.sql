-- admin.sql: Grade lock and admin bulk operations

-- name: LockStudentGrades :one
INSERT INTO grade_locks (student_id, school_id, locked_by, reason)
VALUES ($1, $2, $3, $4)
RETURNING id, locked_at;

-- name: GetActiveLockForStudent :one
SELECT id, reason FROM grade_locks
WHERE student_id = $1 AND school_id = $2 AND is_active = TRUE;

-- name: DeleteGradeLock :exec
DELETE FROM grade_locks WHERE id = $1;

-- name: UpdateStudentGradeLockState :exec
UPDATE students
SET is_grade_locked = $1, lock_reason = $2
WHERE id = $3 AND school_id = $4;

-- name: GetActiveLocks :many
-- All rows in grade_locks are active; inactive rows are deleted on unlock.
SELECT gl.id, gl.student_id, u.first_name, u.last_name, s.student_number,
       gl.reason, gl.locked_at, gl.locked_by
FROM grade_locks gl
JOIN students s ON s.id = gl.student_id
JOIN users u ON u.id = s.user_id
WHERE gl.school_id = $1
ORDER BY gl.locked_at DESC;

-- name: UpdateSchoolSettings :exec
UPDATE schools SET settings = $1, updated_at = NOW() WHERE id = $2;

-- name: GetSchoolSettings :one
SELECT id, name, address, logo_url, settings FROM schools WHERE id = $1;
