-- name: CheckRefreshTokenQuery :one
SELECT
  refresh_token
FROM
  users
WHERE
  email = $1;

-- name: RevokeRefreshTokenQuery :one
UPDATE
  users
SET
  refresh_token = NULL,
  updated_at = NOW()
WHERE
  email = $1
RETURNING
  email;

-- name: CheckEmailExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      users
    WHERE
      email = $1
      AND is_verified = TRUE);

-- name: CheckPendingOnboarding :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      user_onboarding
    WHERE
      email = $1);

-- name: UpsertUserOnboarding :one
INSERT INTO user_onboarding(name, email, password, otp, expires_at, attempts)
  VALUES ($1, $2, $3, $4, $5, 0)
ON CONFLICT (email)
  DO UPDATE SET
    name = EXCLUDED.name,
    password = EXCLUDED.password,
    otp = EXCLUDED.otp,
    expires_at = EXCLUDED.expires_at,
    attempts = 0
  RETURNING
    id,
    email,
    name,
    otp,
    expires_at,
    attempts;