-- 009_create_grade_locks.sql
CREATE TABLE IF NOT EXISTS grade_locks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id   UUID NOT NULL REFERENCES students(id),
    school_id    UUID NOT NULL REFERENCES schools(id),
    locked_by    UUID NOT NULL REFERENCES users(id),
    reason       TEXT NOT NULL,
    locked_at    TIMESTAMPTZ DEFAULT NOW(),
    unlocked_at  TIMESTAMPTZ,
    unlocked_by  UUID REFERENCES users(id),
    is_active    BOOLEAN DEFAULT TRUE
);

CREATE INDEX idx_grade_locks_student ON grade_locks(student_id, is_active);
CREATE INDEX idx_grade_locks_school ON grade_locks(school_id, is_active);
