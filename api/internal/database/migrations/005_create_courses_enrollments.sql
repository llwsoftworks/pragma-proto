-- 005_create_courses_enrollments.sql
CREATE TABLE IF NOT EXISTS courses (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id     UUID NOT NULL REFERENCES schools(id),
    teacher_id    UUID NOT NULL REFERENCES teachers(id),
    name          TEXT NOT NULL,
    subject       TEXT NOT NULL,
    period        TEXT,
    room          TEXT,
    academic_year TEXT NOT NULL,
    semester      TEXT,
    is_active     BOOLEAN DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_courses_school ON courses(school_id);
CREATE INDEX idx_courses_teacher ON courses(teacher_id);

CREATE TABLE IF NOT EXISTS enrollments (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id  UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    course_id   UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    school_id   UUID NOT NULL REFERENCES schools(id),
    enrolled_at TIMESTAMPTZ DEFAULT NOW(),
    dropped_at  TIMESTAMPTZ,
    status      TEXT DEFAULT 'active' CHECK (status IN ('active', 'dropped', 'completed')),
    UNIQUE (student_id, course_id)
);

CREATE INDEX idx_enrollments_student ON enrollments(student_id, status);
CREATE INDEX idx_enrollments_course ON enrollments(course_id, status);
CREATE INDEX idx_enrollments_school ON enrollments(school_id);
