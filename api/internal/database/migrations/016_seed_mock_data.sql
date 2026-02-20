-- 016_seed_mock_data.sql
-- Demo seed data for Lincoln High School.
-- Applied automatically by the migration runner on first start.
-- Safe to re-run: all inserts use ON CONFLICT DO NOTHING.
--
-- ┌──────────────────────────────────────────────────────────────┐
-- │  DEMO CREDENTIALS                                            │
-- │                                                              │
-- │  School ID:  a1b2c3d4-e5f6-7890-abcd-ef1234567890           │
-- │                                                              │
-- │  admin@lincoln.edu        AdminPassword1!   (admin)         │
-- │  j.smith@lincoln.edu      TeacherPass1!#    (teacher)       │
-- │  m.jones@lincoln.edu      TeacherPass2!#    (teacher)       │
-- │  maria.m@lincoln.edu      ParentPass123!    (parent)        │
-- │  sofia.m@lincoln.edu      StudentPass12!    (student)       │
-- │  diego.m@lincoln.edu      StudentPass13!    (student)       │
-- │  alex.k@lincoln.edu       StudentPass14!    (student)       │
-- │  taylor.c@lincoln.edu     StudentPass15!    (student)       │
-- │  jordan.p@lincoln.edu     StudentPass16!    (student)       │
-- └──────────────────────────────────────────────────────────────┘
--
-- RLS note: this migration runs as the table owner, which bypasses
-- RLS by default (ENABLE ROW LEVEL SECURITY without FORCE).
-- The session variable is set anyway for correctness.

SET LOCAL app.current_school_id = 'a1b2c3d4-e5f6-7890-abcd-ef1234567890';

-- ─── School ──────────────────────────────────────────────────────────────────

INSERT INTO schools (id, name, address, settings) VALUES (
    'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
    'Lincoln High School',
    '400 Lincoln Ave, Springfield, IL 62701',
    '{
        "grading_scale": {
            "A":  {"min": 90, "gpa": 4.0},
            "A-": {"min": 88, "gpa": 3.7},
            "B+": {"min": 85, "gpa": 3.3},
            "B":  {"min": 80, "gpa": 3.0},
            "B-": {"min": 78, "gpa": 2.7},
            "C+": {"min": 75, "gpa": 2.3},
            "C":  {"min": 70, "gpa": 2.0},
            "C-": {"min": 68, "gpa": 1.7},
            "D":  {"min": 60, "gpa": 1.0},
            "F":  {"min": 0,  "gpa": 0.0}
        },
        "ai_enabled": true,
        "primary_color": "#1d4ed8",
        "academic_year": "2025-2026"
    }'::jsonb
) ON CONFLICT (id) DO NOTHING;

-- ─── Users ───────────────────────────────────────────────────────────────────
-- Passwords hashed with Argon2id (m=65536, t=3, p=4) per spec §7.1.

INSERT INTO users (id, school_id, role, email, password_hash, first_name, last_name) VALUES

-- Admin
('11111111-1111-1111-1111-111111111111',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'admin', 'admin@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$KSUgvGGz0yEVQU4U1k+p3w$S4bpBDQR8ysaVsE/FGYrdiJxuSXTQmN9ZO8VKmJnLKM',
 'Alex', 'Admin'),

-- Teachers
('22222222-2222-2222-2222-222222222222',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'teacher', 'j.smith@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$khOMDj/6dNpu+HxPNlmA0A$FCUYFpevIWz8I7FJhLqzE4wC02mURo6W1eHzOR+wc8U',
 'Jane', 'Smith'),

('33333333-3333-3333-3333-333333333333',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'teacher', 'm.jones@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$Bm4yVOOokbvFDT3ntDYTbQ$bajLAQJfYq20Y4KKZnGCuVsfeYulQgpnOXMTYWSj+oE',
 'Michael', 'Jones'),

-- Parent
('44444444-4444-4444-4444-444444444444',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'parent', 'maria.m@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$FhQGmqS9ykhbDcQo9+MTiQ$9o9mUndsdzBb4wnKuMxTBA0NdUFuA3HxWQG850QJSWg',
 'Maria', 'Martinez'),

-- Students
('55555555-5555-5555-5555-555555555555',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'student', 'sofia.m@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$HHWpo1neG1amxpdFqoPBYQ$33XG9QYvoeJGYEcEPjQ8SZcf5LbwjwaJoegZBBvhpWc',
 'Sofia', 'Martinez'),

('66666666-6666-6666-6666-666666666666',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'student', 'diego.m@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$y1cyAxUUkIzdWFjxs80PtQ$QKHheWximLDVoZJPG+en+TFSX7/ZyM0ToOzWMbC9GJ8',
 'Diego', 'Martinez'),

('77777777-7777-7777-7777-777777777777',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'student', 'alex.k@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$Zn8yxdjG0jMkri7yM+Hp6g$kt15quj5C6mYjdzw3cSkCxQJoa6N54BzkAbGeRUkcYI',
 'Alex', 'Kim'),

('88888888-8888-8888-8888-888888888888',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'student', 'taylor.c@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$y9QFag8dBbYOfybXXdVWFw$7GGEIs0Yt737iJpWKq/wFKQejSe9iL8/zK3P1T1GH64',
 'Taylor', 'Chen'),

('99999999-9999-9999-9999-999999999999',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'student', 'jordan.p@lincoln.edu',
 '$argon2id$v=19$m=65536,t=3,p=4$m+Bzis+cUYFYXBxPWQNqkQ$etdSPs8SKwH0gDlBBcLzjHcZH3MiZUUf/xDzC8HnvOM',
 'Jordan', 'Patel')

ON CONFLICT (school_id, email) DO NOTHING;

-- ─── Teachers ────────────────────────────────────────────────────────────────

INSERT INTO teachers (id, user_id, school_id, department, title) VALUES
('ffffffff-ffff-ffff-ffff-ffffffffffff',
 '22222222-2222-2222-2222-222222222222',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Mathematics', 'Department Head'),

('abababab-abab-abab-abab-abababababab',
 '33333333-3333-3333-3333-333333333333',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Science', 'Lead Teacher')

ON CONFLICT (user_id) DO NOTHING;

-- ─── Students ─────────────────────────────────────────────────────────────────

INSERT INTO students (id, user_id, school_id, student_number, grade_level, enrollment_date) VALUES
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
 '55555555-5555-5555-5555-555555555555',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'LHS-2026-001', '9th', '2024-08-26'),

('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
 '66666666-6666-6666-6666-666666666666',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'LHS-2026-002', '6th', '2024-08-26'),

('cccccccc-cccc-cccc-cccc-cccccccccccc',
 '77777777-7777-7777-7777-777777777777',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'LHS-2026-003', '9th', '2024-08-26'),

('dddddddd-dddd-dddd-dddd-dddddddddddd',
 '88888888-8888-8888-8888-888888888888',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'LHS-2026-004', '10th', '2023-08-28'),

('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee',
 '99999999-9999-9999-9999-999999999999',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'LHS-2026-005', '9th', '2024-08-26')

ON CONFLICT (user_id) DO NOTHING;

-- ─── Parent → Student links ───────────────────────────────────────────────────
-- Maria Martinez is the parent of Sofia (9th) and Diego (6th).

INSERT INTO parent_students
    (parent_id, student_id, school_id, relationship, is_primary_contact) VALUES
('44444444-4444-4444-4444-444444444444',
 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'mother', TRUE),

('44444444-4444-4444-4444-444444444444',
 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'mother', TRUE)

ON CONFLICT (parent_id, student_id) DO NOTHING;

-- ─── Courses ─────────────────────────────────────────────────────────────────

INSERT INTO courses (id, school_id, teacher_id, name, subject, period, room, academic_year, semester) VALUES
('c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'ffffffff-ffff-ffff-ffff-ffffffffffff',
 'Algebra II', 'Mathematics', 'Period 1', 'Room 101', '2025-2026', 'Full Year'),

('c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'ffffffff-ffff-ffff-ffff-ffffffffffff',
 'Geometry', 'Mathematics', 'Period 3', 'Room 101', '2025-2026', 'Full Year'),

('c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'abababab-abab-abab-abab-abababababab',
 'Biology', 'Science', 'Period 2', 'Room 205', '2025-2026', 'Full Year'),

('c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'abababab-abab-abab-abab-abababababab',
 'Physical Science', 'Science', 'Period 4', 'Room 205', '2025-2026', 'Full Year')

ON CONFLICT (id) DO NOTHING;

-- ─── Enrollments ─────────────────────────────────────────────────────────────

INSERT INTO enrollments (student_id, course_id, school_id) VALUES
-- Sofia: Algebra II, Geometry, Biology
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
-- Diego: Physical Science
('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
-- Alex: Algebra II, Geometry, Biology
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('cccccccc-cccc-cccc-cccc-cccccccccccc', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
-- Taylor: Algebra II, Biology
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('dddddddd-dddd-dddd-dddd-dddddddddddd', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
-- Jordan: Algebra II, Geometry, Biology
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890'),
('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890')

ON CONFLICT (student_id, course_id) DO NOTHING;

-- ─── Assignments ─────────────────────────────────────────────────────────────

INSERT INTO assignments (id, course_id, school_id, title, description, due_date, max_points, category, is_published) VALUES

-- Algebra II (teacher: Jane Smith)
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1',
 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Homework 5 — Quadratic Equations',
 'Practice problems on solving quadratic equations by factoring and the quadratic formula.',
 NOW() - INTERVAL '14 days', 50, 'homework', TRUE),

('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2',
 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Quiz 3 — Polynomials',
 'Short quiz covering polynomial addition, subtraction, and multiplication.',
 NOW() - INTERVAL '10 days', 30, 'quiz', TRUE),

('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3',
 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Test 2 — Chapter 4: Functions & Graphs',
 'Comprehensive test on functions, domain and range, and graph transformations.',
 NOW() - INTERVAL '5 days', 100, 'test', TRUE),

-- Geometry (teacher: Jane Smith)
('a4a4a4a4-a4a4-a4a4-a4a4-a4a4a4a4a4a4',
 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Homework 3 — Triangle Congruence',
 'Proofs and problems on SSS, SAS, ASA, and AAS congruence postulates.',
 NOW() - INTERVAL '12 days', 25, 'homework', TRUE),

('a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5',
 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Homework 4 — Circle Theorems',
 'Apply chord, tangent, and arc theorems to solve geometry problems.',
 NOW() - INTERVAL '6 days', 25, 'homework', TRUE),

('a6a6a6a6-a6a6-a6a6-a6a6-a6a6a6a6a6a6',
 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Quiz 2 — Triangles & Angles',
 'Quiz on angle relationships, triangle sum theorem, and exterior angles.',
 NOW() - INTERVAL '3 days', 30, 'quiz', TRUE),

-- Biology (teacher: Michael Jones)
('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7',
 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Lab Report 1 — Cell Structure',
 'Write a formal lab report on the microscopy lab comparing plant and animal cells.',
 NOW() - INTERVAL '15 days', 100, 'project', TRUE),

('a8a8a8a8-a8a8-a8a8-a8a8-a8a8a8a8a8a8',
 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Quiz 1 — Cells & Organelles',
 'Identify organelles and their functions; cell membrane structure.',
 NOW() - INTERVAL '8 days', 30, 'quiz', TRUE),

('a9a9a9a9-a9a9-a9a9-a9a9-a9a9a9a9a9a9',
 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Midterm Exam',
 'Covers Units 1–4: biochemistry, cell biology, photosynthesis, and cellular respiration.',
 NOW() - INTERVAL '2 days', 100, 'exam', TRUE),

-- Physical Science (teacher: Michael Jones)
('b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1',
 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Homework 1 — Matter & Energy',
 'Reading questions and exercises on states of matter and energy transformations.',
 NOW() - INTERVAL '10 days', 25, 'homework', TRUE),

('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2',
 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4',
 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
 'Quiz 1 — Properties of Matter',
 'Multiple choice and short answer on physical vs. chemical properties.',
 NOW() - INTERVAL '4 days', 20, 'quiz', TRUE)

ON CONFLICT (id) DO NOTHING;

-- ─── Grades ──────────────────────────────────────────────────────────────────
-- graded_by = teacher user id for the course

INSERT INTO grades
    (assignment_id, student_id, school_id, points_earned, letter_grade, graded_by, graded_at) VALUES

-- ── Sofia Martinez (strong student, GPA ≈ 3.6) ──

-- Algebra II
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 47,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '13 days'),
('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 28,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '9 days'),
('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 92,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '4 days'),
-- Geometry
('a4a4a4a4-a4a4-a4a4-a4a4-a4a4a4a4a4a4', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 25,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '11 days'),
('a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 22,  'A-', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '5 days'),
('a6a6a6a6-a6a6-a6a6-a6a6-a6a6a6a6a6a6', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 28,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '2 days'),
-- Biology
('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 97,  'A',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '14 days'),
('a8a8a8a8-a8a8-a8a8-a8a8-a8a8a8a8a8a8', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 29,  'A',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '7 days'),
('a9a9a9a9-a9a9-a9a9-a9a9-a9a9a9a9a9a9', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 88,  'B+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '1 day'),

-- ── Diego Martinez (average student) ──

-- Physical Science
('b1b1b1b1-b1b1-b1b1-b1b1-b1b1b1b1b1b1', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 21,  'B',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '9 days'),
('b2b2b2b2-b2b2-b2b2-b2b2-b2b2b2b2b2b2', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 16,  'B',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '3 days'),

-- ── Alex Kim (above average) ──

-- Algebra II
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 40,  'B',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '13 days'),
('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 22,  'C+', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '9 days'),
('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 78,  'B-', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '4 days'),
-- Geometry
('a4a4a4a4-a4a4-a4a4-a4a4-a4a4a4a4a4a4', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 20,  'B',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '11 days'),
('a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 18,  'C+', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '5 days'),
('a6a6a6a6-a6a6-a6a6-a6a6-a6a6a6a6a6a6', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 23,  'C+', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '2 days'),
-- Biology
('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 88,  'B+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '14 days'),
('a8a8a8a8-a8a8-a8a8-a8a8-a8a8a8a8a8a8', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 25,  'B+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '7 days'),
('a9a9a9a9-a9a9-a9a9-a9a9-a9a9a9a9a9a9', 'cccccccc-cccc-cccc-cccc-cccccccccccc', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 82,  'B',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '1 day'),

-- ── Taylor Chen (excellent student) ──

-- Algebra II
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 48,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '13 days'),
('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 29,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '9 days'),
('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 95,  'A',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '4 days'),
-- Biology
('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 92,  'A',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '14 days'),
('a8a8a8a8-a8a8-a8a8-a8a8-a8a8a8a8a8a8', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 27,  'A',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '7 days'),
('a9a9a9a9-a9a9-a9a9-a9a9-a9a9a9a9a9a9', 'dddddddd-dddd-dddd-dddd-dddddddddddd', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 90,  'A',  '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '1 day'),

-- ── Jordan Patel (declining — triggers AI "needs attention" alert) ──

-- Algebra II (dropping: HW 76% → Quiz 60% → Test 71%)
('a1a1a1a1-a1a1-a1a1-a1a1-a1a1a1a1a1a1', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 38,  'C+', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '13 days'),
('a2a2a2a2-a2a2-a2a2-a2a2-a2a2a2a2a2a2', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 18,  'D',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '9 days'),
('a3a3a3a3-a3a3-a3a3-a3a3-a3a3a3a3a3a3', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 71,  'C',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '4 days'),
-- Geometry
('a4a4a4a4-a4a4-a4a4-a4a4-a4a4a4a4a4a4', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 18,  'C+', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '11 days'),
('a5a5a5a5-a5a5-a5a5-a5a5-a5a5a5a5a5a5', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 15,  'D',  '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '5 days'),
('a6a6a6a6-a6a6-a6a6-a6a6-a6a6a6a6a6a6', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 19,  'C-', '22222222-2222-2222-2222-222222222222', NOW() - INTERVAL '2 days'),
-- Biology
('a7a7a7a7-a7a7-a7a7-a7a7-a7a7a7a7a7a7', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 75,  'C+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '14 days'),
('a8a8a8a8-a8a8-a8a8-a8a8-a8a8a8a8a8a8', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 22,  'C+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '7 days'),
('a9a9a9a9-a9a9-a9a9-a9a9-a9a9a9a9a9a9', 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee', 'a1b2c3d4-e5f6-7890-abcd-ef1234567890', 68,  'D+', '33333333-3333-3333-3333-333333333333', NOW() - INTERVAL '1 day')

ON CONFLICT (assignment_id, student_id) DO NOTHING;

-- ─── Schedule blocks ──────────────────────────────────────────────────────────
-- day_of_week: 1=Mon 2=Tue 3=Wed 4=Thu 5=Fri
-- Courses meet Mon / Wed / Fri (three days per week).

INSERT INTO schedule_blocks
    (school_id, user_id, course_id, day_of_week, start_time, end_time, room, color, semester, is_recurring) VALUES

-- Jane Smith (teacher) — Algebra II (Period 1, Room 101)
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 1, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 3, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 5, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),

-- Jane Smith — Geometry (Period 3, Room 101)
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 1, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 3, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '22222222-2222-2222-2222-222222222222', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 5, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE),

-- Michael Jones (teacher) — Biology (Period 2, Room 205)
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 1, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 3, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 5, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),

-- Michael Jones — Physical Science (Period 4, Room 205)
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4', 1, '11:30', '12:30', 'Room 205', '#d97706', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4', 3, '11:30', '12:30', 'Room 205', '#d97706', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '33333333-3333-3333-3333-333333333333', 'c4c4c4c4-c4c4-c4c4-c4c4-c4c4c4c4c4c4', 5, '11:30', '12:30', 'Room 205', '#d97706', 'Full Year', TRUE),

-- Sofia Martinez (student) — her course schedule
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 1, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 3, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c1c1c1c1-c1c1-c1c1-c1c1-c1c1c1c1c1c1', 5, '08:00', '09:00', 'Room 101', '#1d4ed8', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 1, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 3, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c3c3c3c3-c3c3-c3c3-c3c3-c3c3c3c3c3c3', 5, '09:10', '10:10', 'Room 205', '#059669', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 1, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 3, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE),
('a1b2c3d4-e5f6-7890-abcd-ef1234567890', '55555555-5555-5555-5555-555555555555', 'c2c2c2c2-c2c2-c2c2-c2c2-c2c2c2c2c2c2', 5, '10:20', '11:20', 'Room 101', '#7c3aed', 'Full Year', TRUE)

ON CONFLICT DO NOTHING;
