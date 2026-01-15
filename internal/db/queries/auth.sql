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
      email = $1);

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

-- name: GetPendingOnboardingByEmail :one
SELECT
  id,
  name,
  email,
  PASSWORD,
  otp,
  attempts,
  expires_at
FROM
  user_onboarding
WHERE
  email = $1;

-- name: IncrementOnboardingAttempts :exec
UPDATE user_onboarding
SET attempts = attempts + 1
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users(name, email, password)
  VALUES ($1, $2, $3)
RETURNING
  id, name, email, role, trust_score, created_at, updated_at;

-- name: DeleteOnboardingByEmail :exec
DELETE FROM user_onboarding
WHERE email = $1;

-- name: GetUserByEmail :one
SELECT
  id,
  name,
  email,
  PASSWORD,
  ROLE,
  trust_score,
  created_at,
  updated_at
FROM
  users
WHERE
  email = $1;

-- name: SetUserRefreshToken :exec
UPDATE
  users
SET
  refresh_token = $1,
  updated_at = NOW()
WHERE
  id = $2;

-- name: FetchUserSession :one
SELECT
  id,
  name,
  email,
  ROLE
FROM
  users
WHERE
  email = $1;

-- name: UpsertPasswordReset :one
INSERT INTO password_resets(email, otp, expires_at, attempts)
  VALUES ($1, $2, $3, 0)
ON CONFLICT (email)
  DO UPDATE SET
    otp = EXCLUDED.otp,
    expires_at = EXCLUDED.expires_at,
    attempts = 0
  RETURNING
    id,
    email,
    otp,
    expires_at,
    attempts;

-- name: GetPasswordResetByEmail :one
SELECT
  id,
  email,
  otp,
  attempts,
  expires_at
FROM
  password_resets
WHERE
  email = $1;

-- name: IncrementPasswordResetAttempts :exec
UPDATE password_resets
SET attempts = attempts + 1
WHERE email = $1;

-- name: DeletePasswordResetByEmail :exec
DELETE FROM password_resets
WHERE email = $1;

-- name: UpdateUserPasswordByEmail :exec
UPDATE users
SET password = $2, updated_at = NOW()
WHERE email = $1;
