-- 017_add_short_ids.sql
-- Add short_id columns to tables whose IDs appear in browser URLs.
--
-- short_id is an 8-character base62 string (0-9, A-Z, a-z) used in browser
-- URLs instead of full UUIDs or 22-char base64url encodings. The value is
-- immutable once set, generated at INSERT time by the Go application layer,
-- and looked up via a unique index.
--
-- Tables receiving short_id: courses, assignments.

-- ─── courses ──────────────────────────────────────────────────────────────────

ALTER TABLE courses ADD COLUMN IF NOT EXISTS short_id TEXT;

-- Backfill existing rows with 8 hex chars from md5(uuid). Hex chars (0-9, a-f)
-- are a valid subset of base62. New rows get full base62 IDs from Go.
UPDATE courses SET short_id = substr(md5(id::text), 1, 8)
WHERE short_id IS NULL;

ALTER TABLE courses ALTER COLUMN short_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_courses_short_id ON courses(short_id);

-- ─── assignments ──────────────────────────────────────────────────────────────

ALTER TABLE assignments ADD COLUMN IF NOT EXISTS short_id TEXT;

UPDATE assignments SET short_id = substr(md5(id::text), 1, 8)
WHERE short_id IS NULL;

ALTER TABLE assignments ALTER COLUMN short_id SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_assignments_short_id ON assignments(short_id);
