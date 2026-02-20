-- audit.sql: Audit log queries (INSERT only — no UPDATE or DELETE)

-- name: InsertAuditLog :exec
INSERT INTO audit_logs
    (school_id, user_id, action, entity_type, entity_id,
     old_value, new_value, ip_address, user_agent, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW());

-- name: GetAuditLogsByEntity :many
SELECT id, user_id, action, old_value, new_value, ip_address, created_at
FROM audit_logs
WHERE entity_type = $1 AND entity_id = $2 AND school_id = $3
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAuditLogsByUser :many
SELECT id, action, entity_type, entity_id, old_value, new_value, ip_address, created_at
FROM audit_logs
WHERE user_id = $1 AND school_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: GetRecentAuditLogs :many
SELECT id, user_id, action, entity_type, entity_id, created_at
FROM audit_logs
WHERE school_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
