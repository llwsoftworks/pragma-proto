-- 022_add_short_ids_remaining.sql
-- Add 8-character short_id columns to tables whose UUIDs still appear in URLs.
-- This mirrors migration 017 which added short_ids to courses and assignments.

-- Students: used in /students/{studentId}/grades, /admin/students/{studentId}/lock, etc.
ALTER TABLE students ADD COLUMN IF NOT EXISTS short_id VARCHAR(8);

-- Generate short_ids for existing students using left(md5(id::text), 8).
UPDATE students SET short_id = left(md5(id::text), 8) WHERE short_id IS NULL;

ALTER TABLE students ALTER COLUMN short_id SET NOT NULL;
ALTER TABLE students ALTER COLUMN short_id SET DEFAULT left(md5(gen_random_uuid()::text), 8);
CREATE UNIQUE INDEX IF NOT EXISTS idx_students_short_id ON students(short_id);

-- Schedule blocks: used in DELETE /schedule/{blockId}.
ALTER TABLE schedule_blocks ADD COLUMN IF NOT EXISTS short_id VARCHAR(8);

UPDATE schedule_blocks SET short_id = left(md5(id::text), 8) WHERE short_id IS NULL;

ALTER TABLE schedule_blocks ALTER COLUMN short_id SET NOT NULL;
ALTER TABLE schedule_blocks ALTER COLUMN short_id SET DEFAULT left(md5(gen_random_uuid()::text), 8);
CREATE UNIQUE INDEX IF NOT EXISTS idx_schedule_blocks_short_id ON schedule_blocks(short_id);

-- Digital IDs: used in DELETE /digital-ids/{idId}.
ALTER TABLE digital_ids ADD COLUMN IF NOT EXISTS short_id VARCHAR(8);

UPDATE digital_ids SET short_id = left(md5(id::text), 8) WHERE short_id IS NULL;

ALTER TABLE digital_ids ALTER COLUMN short_id SET NOT NULL;
ALTER TABLE digital_ids ALTER COLUMN short_id SET DEFAULT left(md5(gen_random_uuid()::text), 8);
CREATE UNIQUE INDEX IF NOT EXISTS idx_digital_ids_short_id ON digital_ids(short_id);
