-- name: FetchAllUsers :many
SELECT
  id,
  name,
  email,
  ROLE,
  is_banned,
  trust_score,
  created_at,
  updated_at
FROM
  users
ORDER BY
  created_at DESC;

-- name: FetchUserById :one
SELECT
  id,
  name,
  email,
  ROLE,
  is_banned,
  trust_score,
  created_at,
  updated_at
FROM
  users
WHERE
  id = $1;

-- name: BanUser :execrows
UPDATE
  users
SET
  is_banned = TRUE,
  updated_at = NOW()
WHERE
  id = $1
  AND ROLE != 'ADMIN';

-- name: UnbanUser :exec
UPDATE
  users
SET
  is_banned = FALSE,
  updated_at = NOW()
WHERE
  id = $1;

-- name: PromoteUserToAdmin :execrows
UPDATE
  users
SET
  ROLE = 'ADMIN',
  updated_at = NOW()
WHERE
  id = $1
  AND ROLE = 'USER';

-- name: CheckUserBanned :one
SELECT
  is_banned
FROM
  users
WHERE
  id = $1;
