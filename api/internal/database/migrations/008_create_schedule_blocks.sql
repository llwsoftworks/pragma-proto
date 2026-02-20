-- 008_create_schedule_blocks.sql
CREATE TABLE IF NOT EXISTS schedule_blocks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id    UUID NOT NULL REFERENCES schools(id),
    user_id      UUID NOT NULL REFERENCES users(id),
    course_id    UUID REFERENCES courses(id),
    day_of_week  INT NOT NULL CHECK (day_of_week BETWEEN 0 AND 6),
    start_time   TIME NOT NULL,
    end_time     TIME NOT NULL,
    room         TEXT,
    label        TEXT,
    color        TEXT,
    semester     TEXT,
    is_recurring BOOLEAN DEFAULT TRUE,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    CHECK (end_time > start_time)
);

CREATE INDEX idx_schedule_blocks_user ON schedule_blocks(user_id, day_of_week);
CREATE INDEX idx_schedule_blocks_school ON schedule_blocks(school_id);
CREATE INDEX idx_schedule_blocks_course ON schedule_blocks(course_id);
