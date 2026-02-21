-- 018_seed_super_admin.sql
-- Add the super_admin demo user for testing.
-- Migration 016 was already applied on existing databases before this user
-- was added, so a separate migration ensures the row reaches all environments.
-- Safe to re-run: uses ON CONFLICT DO NOTHING.

INSERT INTO users (id, school_id, role, email, password_hash, first_name, last_name) VALUES
('00000000-0000-0000-0000-000000000001',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'super_admin', 'superadmin@pragma.dev',
 '$argon2id$v=19$m=65536,t=3,p=4$gSLh3SLP5nU6aKypeu5WSA$6nC0VW7mp3rPDXmdZ8eNU4pj2xW65vK7H7rExYZBtdo',
 'Super', 'Admin')
ON CONFLICT (school_id, email) DO NOTHING;
