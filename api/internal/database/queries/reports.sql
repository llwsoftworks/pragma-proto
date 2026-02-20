-- reports.sql: Report card and document queries

-- name: CreateReportCard :one
INSERT INTO report_cards
    (student_id, school_id, academic_period, gpa, teacher_comments, admin_comments,
     is_finalized, pdf_url, generated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, generated_at;

-- name: GetReportCardsByStudent :many
SELECT id, academic_period, gpa, is_finalized, pdf_url, generated_at
FROM report_cards
WHERE student_id = $1 AND school_id = $2
ORDER BY generated_at DESC;

-- name: FinalizeReportCard :exec
UPDATE report_cards SET is_finalized = TRUE WHERE id = $1 AND school_id = $2;

-- name: CreateDocument :one
INSERT INTO documents
    (id, school_id, student_id, type, verification_code, pdf_url, generated_by, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at;

-- name: GetDocumentByVerificationCode :one
SELECT d.id, d.type, d.student_id, d.expires_at, d.created_at,
       u.first_name, u.last_name
FROM documents d
JOIN students s ON s.id = d.student_id
JOIN users u ON u.id = s.user_id
WHERE d.verification_code = $1
LIMIT 1;

-- name: GetDocumentsByStudent :many
SELECT id, type, verification_code, pdf_url, expires_at, created_at
FROM documents
WHERE student_id = $1 AND school_id = $2
ORDER BY created_at DESC;

-- name: GetDailyDocumentCount :one
SELECT COUNT(*)::int
FROM documents
WHERE generated_by = $1
  AND created_at >= NOW()::date
  AND created_at < NOW()::date + INTERVAL '1 day';
