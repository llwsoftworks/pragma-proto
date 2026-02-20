-- assignments.sql: Assignment CRUD and attachment queries

-- name: CreateAssignment :one
INSERT INTO assignments
    (course_id, school_id, title, description, due_date,
     max_points, category, weight, is_published)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, created_at;

-- name: GetAssignmentByID :one
SELECT id, course_id, school_id, title, description, due_date,
       max_points, category, weight, is_published, created_at, updated_at
FROM assignments
WHERE id = $1 AND school_id = $2
LIMIT 1;

-- name: UpdateAssignment :exec
UPDATE assignments
SET title = $1, description = $2, due_date = $3, max_points = $4,
    category = $5, weight = $6, is_published = $7, updated_at = NOW()
WHERE id = $8 AND school_id = $9;

-- name: DeleteAssignment :exec
DELETE FROM assignments WHERE id = $1 AND school_id = $2;

-- name: ListAssignmentsByCourse :many
SELECT id, title, description, due_date, max_points, category, weight, is_published, created_at
FROM assignments
WHERE course_id = $1 AND school_id = $2
ORDER BY due_date NULLS LAST, created_at;

-- name: ListPublishedAssignmentsByCourse :many
SELECT id, title, description, due_date, max_points, category, weight, created_at
FROM assignments
WHERE course_id = $1 AND school_id = $2 AND is_published = TRUE
ORDER BY due_date NULLS LAST, created_at;

-- name: CreateAttachment :one
INSERT INTO assignment_attachments
    (assignment_id, school_id, file_name, file_key, file_size, mime_type, uploaded_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: VersionReplaceAttachment :exec
-- Mark old attachments as not current when a file is replaced.
UPDATE assignment_attachments
SET is_current = FALSE
WHERE assignment_id = $1 AND school_id = $2 AND file_name = $3 AND is_current = TRUE;

-- name: GetCurrentAttachments :many
SELECT id, file_name, file_key, file_size, mime_type, version, created_at
FROM assignment_attachments
WHERE assignment_id = $1 AND school_id = $2 AND is_current = TRUE
ORDER BY created_at DESC;
