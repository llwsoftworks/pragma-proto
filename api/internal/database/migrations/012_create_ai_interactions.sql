-- 012_create_ai_interactions.sql
CREATE TABLE IF NOT EXISTS ai_interactions (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id      UUID NOT NULL REFERENCES schools(id),
    user_id        UUID NOT NULL REFERENCES users(id),
    feature        TEXT NOT NULL CHECK (feature IN ('grading_assistant', 'student_insights', 'report_comments', 'smart_scheduling', 'parent_communication')),
    input_summary  TEXT,
    output_summary TEXT,
    tokens_used    INT,
    accepted       BOOLEAN,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ai_interactions_school ON ai_interactions(school_id, created_at DESC);
CREATE INDEX idx_ai_interactions_user ON ai_interactions(user_id, created_at DESC);
