-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY id;

-- Sessions

-- name: CreateSession :one
INSERT INTO sessions (user_id, token, expires_at, user_agent, ip_address)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions
WHERE token = $1;

-- name: ListActiveSessionsByUser :many
SELECT * FROM sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: RevokeSessionByToken :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE token = $1
  AND revoked_at IS NULL;

-- name: RevokeAllSessionsByUser :exec
UPDATE sessions
SET revoked_at = NOW()
WHERE user_id = $1
  AND revoked_at IS NULL;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at <= NOW();
