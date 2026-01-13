-- name: CheckRefreshTokenQuery :one
SELECT refresh_token 
FROM users 
WHERE email = $1;

-- name: RevokeRefreshTokenQuery :one
UPDATE users
SET
  refresh_token = NULL,
  updated_at = NOW()
WHERE
  email = $1
RETURNING
  email;