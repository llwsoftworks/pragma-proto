-- 021_fix_rls_for_super_admins.sql
-- Update the RLS policy on users to allow super_admin rows with NULL school_id
-- to be visible (they are not scoped to any school).

DROP POLICY IF EXISTS tenant_isolation_users ON users;

CREATE POLICY tenant_isolation_users ON users
    USING (
        school_id = current_setting('app.current_school_id', TRUE)::UUID
        OR school_id IS NULL  -- super_admin rows are visible from any tenant context
    );

-- Sessions can also have NULL school_id for super_admin sessions.
DROP POLICY IF EXISTS tenant_isolation_sessions ON sessions;

CREATE POLICY tenant_isolation_sessions ON sessions
    USING (
        school_id = current_setting('app.current_school_id', TRUE)::UUID
        OR school_id IS NULL
    );
