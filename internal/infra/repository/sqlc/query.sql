-- name: GetUser :one
SELECT *
FROM users
WHERE username = ? LIMIT 1;

-- name: GetUsers :many
SELECT id, username, display_name, role, created_at, updated_at
FROM users;

-- name: GetUserById :one
SELECT *
FROM users
WHERE id = ? LIMIT 1;

-- name: CreateUser :execresult
INSERT INTO users (username, password, display_name, role, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;