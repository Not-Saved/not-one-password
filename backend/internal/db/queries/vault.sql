-- name: GetVaultByUserID :one
SELECT *
FROM vaults
WHERE user_id = $1;

-- name: GetVaultUpdatedAtByUserID :one
SELECT updated_at
FROM vaults
WHERE user_id=$1;

-- name: InsertVaultByUserID :one
INSERT INTO vaults (user_id, vault)
VALUES ($1, $2)
ON CONFLICT (user_id)
DO UPDATE SET
    vault = EXCLUDED.vault,
    updated_at = NOW()
RETURNING *;
