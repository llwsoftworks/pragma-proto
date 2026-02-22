-- 019_cleanup_grade_locks.sql
-- Remove stale grade_lock rows that remain after students are unlocked.
-- Going forward, unlock deletes the row entirely. The audit_logs table
-- preserves the full lock/unlock history so these rows are unnecessary bloat.

-- 1. Delete all inactive (already-unlocked) grade_lock rows.
DELETE FROM grade_locks WHERE is_active = FALSE;

-- 2. Drop columns that are no longer needed now that unlocked rows are deleted.
ALTER TABLE grade_locks DROP COLUMN IF EXISTS unlocked_at;
ALTER TABLE grade_locks DROP COLUMN IF EXISTS unlocked_by;
ALTER TABLE grade_locks DROP COLUMN IF EXISTS is_active;

-- 3. Re-create a simpler index (the old composite index on is_active is gone).
DROP INDEX IF EXISTS idx_grade_locks_student;
DROP INDEX IF EXISTS idx_grade_locks_school;

CREATE INDEX idx_grade_locks_student ON grade_locks(student_id);
CREATE INDEX idx_grade_locks_school ON grade_locks(school_id);
