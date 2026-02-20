-- 013_create_audit_logs.sql
-- Append-only: no UPDATE or DELETE permissions granted to the application role.
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id   UUID NOT NULL REFERENCES schools(id),
    user_id     UUID REFERENCES users(id),
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   UUID,
    old_value   JSONB,
    new_value   JSONB,
    ip_address  INET,
    user_agent  TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, created_at DESC);
CREATE INDEX idx_audit_logs_school ON audit_logs(school_id, created_at DESC);

-- Revoke mutation permissions from the application database role.
-- The application role must be configured with: REVOKE UPDATE, DELETE ON audit_logs FROM app_role;
-- This is enforced at the PostgreSQL level as defense-in-depth.
