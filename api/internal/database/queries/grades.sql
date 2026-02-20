-- grades.sql: Grade CRUD and aggregation queries

-- name: UpsertGrade :one
INSERT INTO grades
    (assignment_id, student_id, school_id, points_earned, comment,
     is_excused, is_missing, is_late, ai_accepted, graded_by, graded_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
ON CONFLICT (assignment_id, student_id)
DO UPDATE SET
    points_earned = EXCLUDED.points_earned,
    comment       = EXCLUDED.comment,
    is_excused    = EXCLUDED.is_excused,
    is_missing    = EXCLUDED.is_missing,
    is_late       = EXCLUDED.is_late,
    ai_accepted   = EXCLUDED.ai_accepted,
    graded_by     = EXCLUDED.graded_by,
    graded_at     = NOW(),
    updated_at    = NOW()
RETURNING id, updated_at;

-- name: GetGradesByCourse :many
SELECT g.id, g.assignment_id, g.student_id, g.points_earned,
       g.letter_grade, g.comment, g.is_excused, g.is_missing, g.is_late,
       g.ai_suggested, g.ai_accepted, g.graded_at, g.updated_at
FROM grades g
JOIN assignments a ON a.id = g.assignment_id
WHERE a.course_id = $1 AND g.school_id = $2
ORDER BY g.student_id, a.due_date NULLS LAST;

-- name: GetGradesByStudent :many
SELECT g.id, g.assignment_id, g.student_id, g.points_earned,
       g.letter_grade, g.comment, g.is_excused, g.is_missing, g.is_late,
       a.title, a.max_points, a.category, a.due_date, c.name AS course_name
FROM grades g
JOIN assignments a ON a.id = g.assignment_id
JOIN courses c ON c.id = a.course_id
WHERE g.student_id = $1 AND g.school_id = $2
  AND a.is_published = TRUE
ORDER BY c.name, a.due_date NULLS LAST;

-- name: GetUngradedCount :one
SELECT COUNT(*)::int
FROM assignments a
JOIN courses c ON c.id = a.course_id
JOIN teachers t ON t.id = c.teacher_id
JOIN enrollments e ON e.course_id = a.course_id AND e.status = 'active'
WHERE t.user_id = $1 AND a.school_id = $2
  AND a.due_date < NOW()
  AND NOT EXISTS (
    SELECT 1 FROM grades g
    WHERE g.assignment_id = a.id AND g.student_id = e.student_id
  );

-- name: StoreAISuggestion :exec
UPDATE grades SET ai_suggested = $1 WHERE assignment_id = $2 AND student_id = $3 AND school_id = $4;

-- name: GetCourseGradeSummary :many
-- Returns average percentage per assignment for the teacher dashboard.
SELECT a.id, a.title, a.max_points,
       COALESCE(AVG(g.points_earned / a.max_points * 100), 0)::float AS avg_percent,
       COUNT(g.id)::int AS graded_count
FROM assignments a
LEFT JOIN grades g ON g.assignment_id = a.id
WHERE a.course_id = $1 AND a.school_id = $2
GROUP BY a.id, a.title, a.max_points
ORDER BY a.updated_at DESC
LIMIT $3;
