-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByPublicID :one
SELECT * FROM users
WHERE public_id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY id;
