-- 004_create_parent_students.sql
CREATE TABLE IF NOT EXISTS parent_students (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    student_id          UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    school_id           UUID NOT NULL REFERENCES schools(id),
    relationship        TEXT NOT NULL CHECK (relationship IN ('mother', 'father', 'guardian', 'other')),
    is_primary_contact  BOOLEAN DEFAULT FALSE,
    can_view_grades     BOOLEAN DEFAULT TRUE,
    can_generate_docs   BOOLEAN DEFAULT TRUE,
    created_at          TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (parent_id, student_id)
);

CREATE INDEX idx_parent_students_parent ON parent_students(parent_id);
CREATE INDEX idx_parent_students_student ON parent_students(student_id);
CREATE INDEX idx_parent_students_school ON parent_students(school_id);
