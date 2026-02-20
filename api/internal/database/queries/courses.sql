-- courses.sql: Course CRUD and enrollment queries

-- name: CreateCourse :one
INSERT INTO courses
    (school_id, teacher_id, name, subject, period, room, academic_year, semester)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;

-- name: GetCourseByID :one
SELECT c.id, c.school_id, c.teacher_id, c.name, c.subject,
       c.period, c.room, c.academic_year, c.semester, c.is_active,
       u.first_name || ' ' || u.last_name AS teacher_name
FROM courses c
JOIN teachers t ON t.id = c.teacher_id
JOIN users u ON u.id = t.user_id
WHERE c.id = $1 AND c.school_id = $2
LIMIT 1;

-- name: ListCoursesByTeacher :many
SELECT c.id, c.name, c.subject, c.period, c.room, c.academic_year, c.semester, c.is_active,
       (SELECT COUNT(*)::int FROM enrollments e WHERE e.course_id = c.id AND e.status = 'active') AS enrollment_count
FROM courses c
JOIN teachers t ON t.id = c.teacher_id
WHERE t.user_id = $1 AND c.school_id = $2
ORDER BY c.name;

-- name: ListCoursesBySchool :many
SELECT c.id, c.name, c.subject, c.period, c.academic_year, c.is_active,
       u.first_name || ' ' || u.last_name AS teacher_name,
       (SELECT COUNT(*)::int FROM enrollments e WHERE e.course_id = c.id AND e.status = 'active') AS enrollment_count
FROM courses c
JOIN teachers t ON t.id = c.teacher_id
JOIN users u ON u.id = t.user_id
WHERE c.school_id = $1
ORDER BY c.name;

-- name: EnrollStudent :exec
INSERT INTO enrollments (student_id, course_id, school_id)
VALUES ($1, $2, $3)
ON CONFLICT (student_id, course_id) DO UPDATE SET status = 'active', dropped_at = NULL;

-- name: DropStudentFromCourse :exec
UPDATE enrollments
SET status = 'dropped', dropped_at = NOW()
WHERE student_id = $1 AND course_id = $2 AND school_id = $3;

-- name: GetEnrolledStudents :many
SELECT s.id, u.first_name, u.last_name, s.student_number, s.grade_level
FROM enrollments e
JOIN students s ON s.id = e.student_id
JOIN users u ON u.id = s.user_id
WHERE e.course_id = $1 AND e.status = 'active'
ORDER BY u.last_name, u.first_name;

-- name: GetStudentCourses :many
SELECT c.id, c.name, c.subject, c.period, c.room,
       u.first_name || ' ' || u.last_name AS teacher_name
FROM enrollments e
JOIN courses c ON c.id = e.course_id
JOIN teachers t ON t.id = c.teacher_id
JOIN users u ON u.id = t.user_id
WHERE e.student_id = $1 AND e.status = 'active' AND c.school_id = $2
ORDER BY c.name;
