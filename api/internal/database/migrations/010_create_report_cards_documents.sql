-- 010_create_report_cards_documents.sql
CREATE TABLE IF NOT EXISTS report_cards (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id       UUID NOT NULL REFERENCES students(id),
    school_id        UUID NOT NULL REFERENCES schools(id),
    academic_period  TEXT NOT NULL,
    gpa              DECIMAL(4,3),
    teacher_comments TEXT,
    admin_comments   TEXT,
    is_finalized     BOOLEAN DEFAULT FALSE,
    pdf_url          TEXT,
    generated_by     UUID NOT NULL REFERENCES users(id),
    generated_at     TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_report_cards_student ON report_cards(student_id, school_id);
CREATE INDEX idx_report_cards_school ON report_cards(school_id);

CREATE TABLE IF NOT EXISTS documents (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    school_id         UUID NOT NULL REFERENCES schools(id),
    student_id        UUID NOT NULL REFERENCES students(id),
    type              TEXT NOT NULL CHECK (type IN ('enrollment_certificate', 'attendance_letter', 'academic_standing', 'tuition_confirmation', 'custom')),
    verification_code TEXT NOT NULL UNIQUE,
    pdf_url           TEXT NOT NULL,
    generated_by      UUID NOT NULL REFERENCES users(id),
    expires_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_documents_verification ON documents(verification_code);
CREATE INDEX idx_documents_student ON documents(student_id, school_id);
