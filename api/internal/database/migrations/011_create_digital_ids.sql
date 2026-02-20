-- 011_create_digital_ids.sql
CREATE TABLE IF NOT EXISTS digital_ids (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id    UUID NOT NULL REFERENCES students(id),
    school_id     UUID NOT NULL REFERENCES schools(id),
    id_number     TEXT NOT NULL,
    qr_code_data  TEXT NOT NULL,
    barcode_data  TEXT,
    photo_url     TEXT,
    issued_at     TIMESTAMPTZ DEFAULT NOW(),
    expires_at    TIMESTAMPTZ NOT NULL,
    is_valid      BOOLEAN DEFAULT TRUE,
    revoked_at    TIMESTAMPTZ,
    UNIQUE (school_id, id_number)
);

CREATE INDEX idx_digital_ids_student ON digital_ids(student_id, is_valid);
CREATE INDEX idx_digital_ids_school ON digital_ids(school_id);
