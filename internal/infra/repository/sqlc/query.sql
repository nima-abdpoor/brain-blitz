-- name: GetUser :one
SELECT *
FROM user
WHERE id = ? LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO user (email, password, display_name, created_at, updated_at)
VALUES (?, ?, ?, ?, ?);

-- name: DeleteUser :exec
DELETE FROM user
WHERE id = ?;