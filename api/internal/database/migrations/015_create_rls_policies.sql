-- 015_create_rls_policies.sql
-- PostgreSQL Row Level Security: defense-in-depth tenant isolation.
-- Even if Go middleware has a bug, Postgres blocks cross-tenant data access.

ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE students ENABLE ROW LEVEL SECURITY;
ALTER TABLE teachers ENABLE ROW LEVEL SECURITY;
ALTER TABLE parent_students ENABLE ROW LEVEL SECURITY;
ALTER TABLE courses ENABLE ROW LEVEL SECURITY;
ALTER TABLE enrollments ENABLE ROW LEVEL SECURITY;
ALTER TABLE assignments ENABLE ROW LEVEL SECURITY;
ALTER TABLE assignment_attachments ENABLE ROW LEVEL SECURITY;
ALTER TABLE grades ENABLE ROW LEVEL SECURITY;
ALTER TABLE schedule_blocks ENABLE ROW LEVEL SECURITY;
ALTER TABLE grade_locks ENABLE ROW LEVEL SECURITY;
ALTER TABLE report_cards ENABLE ROW LEVEL SECURITY;
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE digital_ids ENABLE ROW LEVEL SECURITY;
ALTER TABLE sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_interactions ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;

-- The Go API sets this session variable before each query:
-- SET LOCAL app.current_school_id = '<uuid>';

CREATE POLICY tenant_isolation_users ON users
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_students ON students
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_teachers ON teachers
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_parent_students ON parent_students
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_courses ON courses
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_enrollments ON enrollments
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_assignments ON assignments
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_assignment_attachments ON assignment_attachments
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_grades ON grades
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_schedule_blocks ON schedule_blocks
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_grade_locks ON grade_locks
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_report_cards ON report_cards
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_documents ON documents
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_digital_ids ON digital_ids
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_sessions ON sessions
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_ai_interactions ON ai_interactions
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);

CREATE POLICY tenant_isolation_audit_logs ON audit_logs
    USING (school_id = current_setting('app.current_school_id', TRUE)::UUID);
