-- name: GetUserByUsername :one
SELECT * FROM users
WHERE LOWER(username) = LOWER(?) LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
  username, password_hash
) VALUES (
    ?, ?
)
RETURNING *;
