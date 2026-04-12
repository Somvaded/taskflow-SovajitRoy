-- name: CreateUser :exec
INSERT INTO users (id, name, email, password_hash, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at
FROM users
WHERE email = $1;

-- name: GetUserByID :one
SELECT id, name, email, password_hash, created_at
FROM users
WHERE id = $1;
