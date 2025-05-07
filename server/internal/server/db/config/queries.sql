-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE LOWER(username) = LOWER(?)
LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  id, username, password_hash
) VALUES (
  ?, ?, ?
)
RETURNING *;

-- name: SaveRefreshToken :exec
INSERT INTO refresh_tokens (
  jti, user_id, expire_at
) VALUES (
  ?, ?, ?
);

-- name: IsRefreshTokenValid :one
SELECT 1
FROM refresh_tokens
WHERE jti = ?
  AND user_id = ?
  AND revoked_at IS NULL
  AND expired_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE jti = ?;

-- name: RevokeTokensForUser :execrows
UPDATE refresh_tokens
SET revoked_at = CURRENT_TIMESTAMP
WHERE user_id = ?;

-- name: ListActiveTokensForUser :many
SELECT *
FROM refresh_tokens
WHERE user_id = ?;

-- name: DeleteExpiredOrRevokedTokens :execrows
DELETE FROM refresh_tokens
WHERE expire_at <= CURRENT_TIMESTAMP
  OR revoked_at IS NOT NULL;