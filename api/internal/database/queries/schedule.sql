-- schedule.sql: Schedule block CRUD and conflict detection

-- name: GetScheduleBlocksByUser :many
SELECT sb.id, sb.course_id, c.name AS course_name,
       sb.day_of_week, sb.start_time::text, sb.end_time::text,
       sb.room, sb.label, sb.color, sb.is_recurring, sb.semester
FROM schedule_blocks sb
LEFT JOIN courses c ON c.id = sb.course_id
WHERE sb.user_id = $1 AND sb.school_id = $2
ORDER BY sb.day_of_week, sb.start_time;

-- name: CreateScheduleBlock :one
INSERT INTO schedule_blocks
    (school_id, user_id, course_id, day_of_week, start_time, end_time,
     room, label, color, semester, is_recurring)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id;

-- name: DeleteScheduleBlock :exec
DELETE FROM schedule_blocks WHERE id = $1 AND user_id = $2 AND school_id = $3;

-- name: DetectRoomConflicts :many
-- Returns blocks in the same room at an overlapping time, same day.
SELECT sb.id, sb.user_id, sb.course_id, c.name AS course_name,
       sb.start_time::text, sb.end_time::text, sb.label
FROM schedule_blocks sb
LEFT JOIN courses c ON c.id = sb.course_id
WHERE sb.school_id = $1
  AND sb.room = $2
  AND sb.day_of_week = $3
  AND sb.start_time < $5::time  -- new end > existing start
  AND sb.end_time   > $4::time  -- new start < existing end
  AND sb.id != $6;              -- exclude the block being saved (for updates)

-- name: GetUserScheduleForDay :many
SELECT sb.id, sb.course_id, COALESCE(c.name, sb.label, 'Free') AS label,
       sb.start_time::text, sb.end_time::text, sb.room, sb.color
FROM schedule_blocks sb
LEFT JOIN courses c ON c.id = sb.course_id
WHERE sb.user_id = $1 AND sb.day_of_week = $2 AND sb.school_id = $3
ORDER BY sb.start_time;
