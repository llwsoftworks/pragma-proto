-- 006_create_assignments_attachments.sql
CREATE TABLE IF NOT EXISTS assignments (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id    UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    school_id    UUID NOT NULL REFERENCES schools(id),
    title        TEXT NOT NULL,
    description  TEXT,
    due_date     TIMESTAMPTZ,
    max_points   DECIMAL(8,2) NOT NULL,
    category     TEXT NOT NULL CHECK (category IN ('homework', 'quiz', 'test', 'exam', 'project', 'classwork', 'participation', 'other')),
    weight       DECIMAL(5,4) DEFAULT 1.0,
    is_published BOOLEAN DEFAULT FALSE,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_assignments_course ON assignments(course_id, is_published);
CREATE INDEX idx_assignments_school ON assignments(school_id);

CREATE TRIGGER assignments_updated_at
    BEFORE UPDATE ON assignments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TABLE IF NOT EXISTS assignment_attachments (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    school_id     UUID NOT NULL REFERENCES schools(id),
    file_name     TEXT NOT NULL,
    file_key      TEXT NOT NULL,
    file_size     BIGINT NOT NULL,
    mime_type     TEXT NOT NULL,
    uploaded_by   UUID NOT NULL REFERENCES users(id),
    version       INT DEFAULT 1,
    is_current    BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_attachments_assignment ON assignment_attachments(assignment_id, is_current);
