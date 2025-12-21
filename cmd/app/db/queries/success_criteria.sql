-- name: CreateSuccessCriteria :one
INSERT INTO success_criteria (goal_id, user_id, description, completed, position, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSuccessCriteria :one
SELECT * FROM success_criteria
WHERE id = ? AND user_id = ?;

-- name: GetAllSuccessCriteriaByGoal :many
SELECT * FROM success_criteria
WHERE goal_id = ? AND user_id = ?
ORDER BY position ASC, created_at ASC;

-- name: UpdateSuccessCriteria :execresult
UPDATE success_criteria
SET description = ?, completed = ?, position = ?
WHERE id = ? AND user_id = ?;

-- name: ToggleSuccessCriteriaCompleted :execresult
UPDATE success_criteria
SET completed = CASE WHEN completed = 0 THEN 1 ELSE 0 END
WHERE id = ? AND user_id = ?;

-- name: DeleteSuccessCriteria :execresult
DELETE FROM success_criteria
WHERE id = ? AND user_id = ?;

-- name: DeleteAllSuccessCriteriaByGoal :execresult
DELETE FROM success_criteria
WHERE goal_id = ? AND user_id = ?;
