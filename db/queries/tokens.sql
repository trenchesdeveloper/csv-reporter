-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (hashed_token,
                            user_id,
                            expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRefreshToken :one
SELECT *
FROM refresh_tokens
WHERE hashed_token = $1
  AND expires_at > NOW()
LIMIT 1;

-- name: DeleteRefreshToken :exec
DELETE
FROM refresh_tokens
WHERE hashed_token = $1;

-- name: DeleteExpiredRefreshTokens :exec
DELETE
FROM refresh_tokens
WHERE expires_at <= NOW();

-- name: DeleteUserRefreshToken :exec
DELETE FROM refresh_tokens
WHERE user_id = $1
  AND hashed_token = $2;

-- name: UpdateRefreshTokenExpiry :one
UPDATE refresh_tokens
SET expires_at = $2
WHERE hashed_token = $1
RETURNING *;

-- name: GetTokenByPrimaryKey :one
SELECT *
FROM refresh_tokens
WHERE user_id = $1
  AND hashed_token = $2;


-- name: DeleteAllUserRefreshTokens :exec
DELETE FROM refresh_tokens
WHERE user_id = $1;
