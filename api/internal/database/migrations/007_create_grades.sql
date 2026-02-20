-- 007_create_grades.sql
CREATE TABLE IF NOT EXISTS grades (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    assignment_id UUID NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    student_id    UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    school_id     UUID NOT NULL REFERENCES schools(id),
    points_earned DECIMAL(8,2),
    letter_grade  TEXT,
    comment       TEXT,
    graded_by     UUID REFERENCES users(id),
    graded_at     TIMESTAMPTZ,
    ai_suggested  DECIMAL(8,2),
    ai_accepted   BOOLEAN,
    is_excused    BOOLEAN DEFAULT FALSE,
    is_missing    BOOLEAN DEFAULT FALSE,
    is_late       BOOLEAN DEFAULT FALSE,
    created_at    TIMESTAMPTZ DEFAULT NOW(),
    updated_at    TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (assignment_id, student_id)
);

CREATE INDEX idx_grades_school ON grades(school_id);
CREATE INDEX idx_grades_student ON grades(student_id, school_id);
CREATE INDEX idx_grades_assignment ON grades(assignment_id);

CREATE TRIGGER grades_updated_at
    BEFORE UPDATE ON grades
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
