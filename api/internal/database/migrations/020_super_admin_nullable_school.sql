-- 020_super_admin_nullable_school.sql
-- Super-admin users operate across all schools and must not be tied to a
-- single school_id.  Make school_id nullable and relax the UNIQUE constraint
-- so that super_admins can exist without a school.

-- 1. Allow NULL school_id on the users table.
ALTER TABLE users ALTER COLUMN school_id DROP NOT NULL;

-- 2. The existing UNIQUE(school_id, email) breaks for NULLs (Postgres treats
--    NULLs as distinct in unique indexes).  Replace it with a partial unique
--    index for school-bound users, plus a unique index on email alone for
--    super_admins (who have school_id IS NULL).
DROP INDEX IF EXISTS users_school_id_email_key;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_school_id_email_key;

CREATE UNIQUE INDEX idx_users_school_email
    ON users(school_id, email) WHERE school_id IS NOT NULL;

CREATE UNIQUE INDEX idx_users_superadmin_email
    ON users(email) WHERE school_id IS NULL;

-- 3. Clear the placeholder school_id from existing super_admin rows.
UPDATE users SET school_id = NULL WHERE role = 'super_admin';

-- 4. The sessions table also references school_id.  Make it nullable for
--    super_admin sessions.
ALTER TABLE sessions ALTER COLUMN school_id DROP NOT NULL;
