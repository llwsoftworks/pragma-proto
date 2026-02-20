-- dashboard.sql: Dashboard aggregation queries

-- name: GetTeacherDashboard :one
SELECT
    (SELECT COUNT(*)::int FROM courses c JOIN teachers t ON t.id = c.teacher_id
     WHERE t.user_id = $1 AND c.school_id = $2 AND c.is_active = TRUE) AS active_courses,
    (SELECT COUNT(*)::int FROM assignments a
     JOIN courses c ON c.id = a.course_id
     JOIN teachers t ON t.id = c.teacher_id
     WHERE t.user_id = $1 AND a.school_id = $2
       AND a.due_date < NOW()
       AND EXISTS (
           SELECT 1 FROM enrollments e WHERE e.course_id = a.course_id AND e.status = 'active'
           AND NOT EXISTS (SELECT 1 FROM grades g WHERE g.assignment_id = a.id AND g.student_id = e.student_id)
       )) AS ungraded_assignments;

-- name: GetAdminDashboard :one
SELECT
    (SELECT COUNT(*)::int FROM students WHERE school_id = $1 AND enrollment_status = 'active') AS total_students,
    (SELECT COUNT(*)::int FROM teachers WHERE school_id = $1) AS total_teachers,
    (SELECT COUNT(*)::int FROM students WHERE school_id = $1 AND is_grade_locked = TRUE) AS locked_students,
    (SELECT COUNT(*)::int FROM courses WHERE school_id = $1 AND is_active = TRUE) AS active_courses;

-- name: GetStudentGradeSummary :many
-- For parent/student dashboard: current grade per course.
SELECT c.id, c.name AS course_name, c.subject,
       COALESCE(AVG(g.points_earned / a.max_points * 100), 0)::float AS avg_percent
FROM courses c
JOIN enrollments e ON e.course_id = c.id AND e.status = 'active'
JOIN assignments a ON a.course_id = c.id AND a.is_published = TRUE
LEFT JOIN grades g ON g.assignment_id = a.id AND g.student_id = $1
WHERE e.student_id = $1 AND c.school_id = $2
GROUP BY c.id, c.name, c.subject
ORDER BY c.name;
