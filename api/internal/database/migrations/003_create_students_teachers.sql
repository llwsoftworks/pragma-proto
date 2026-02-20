-- 003_create_students_teachers.sql
CREATE TABLE IF NOT EXISTS students (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id         UUID NOT NULL REFERENCES schools(id),
    student_number    TEXT NOT NULL,
    grade_level       TEXT NOT NULL,
    enrollment_date   DATE NOT NULL,
    enrollment_status TEXT DEFAULT 'active' CHECK (enrollment_status IN ('active', 'withdrawn', 'graduated', 'transferred')),
    is_grade_locked   BOOLEAN DEFAULT FALSE,
    lock_reason       TEXT,
    date_of_birth     TEXT, -- AES-256-GCM encrypted
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (school_id, student_number)
);

CREATE INDEX idx_students_school ON students(school_id);
CREATE INDEX idx_students_user ON students(user_id);
CREATE INDEX idx_students_grade_level ON students(school_id, grade_level);

CREATE TABLE IF NOT EXISTS teachers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    school_id   UUID NOT NULL REFERENCES schools(id),
    department  TEXT,
    title       TEXT,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_teachers_school ON teachers(school_id);
CREATE INDEX idx_teachers_user ON teachers(user_id);
