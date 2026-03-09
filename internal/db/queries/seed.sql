-- name: SeedUser :one
INSERT INTO users (name, email, password, role, trust_score)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (email) DO UPDATE SET name = EXCLUDED.name
RETURNING id;

-- name: SeedHub :one
INSERT INTO hubs (name, address, latitude, longitude, contact)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: SeedItem :one
INSERT INTO items (
    user_id, hub_id, name, description, type, status, 
    location_description, latitude, longitude, time_at, is_anonymous
)
VALUES ($1, sqlc.narg('hub_id'), $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, type;

-- name: SeedClaim :one
INSERT INTO claims (item_id, claimant_id, lost_item_id, status, proof_text, similarity_score)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;

-- name: SeedAuditLog :exec
INSERT INTO audit_logs (actor_id, action, target_type, target_id)
VALUES ($1, $2, $3, $4);

-- name: TruncateTables :exec
TRUNCATE TABLE audit_logs, claims, items, hubs, users, user_onboarding, password_resets RESTART IDENTITY CASCADE;
