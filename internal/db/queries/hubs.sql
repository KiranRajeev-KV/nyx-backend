-- name: CreateHub :one
INSERT INTO hubs (name, address, longitude, latitude, contact, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
RETURNING id, name, address, longitude, latitude, contact, created_at, updated_at;

-- name: FetchAllHubs :many
SELECT id, name, address, longitude, latitude, contact, created_at, updated_at
FROM hubs
ORDER BY created_at DESC;

-- name: FetchHubByID :one
SELECT id, name, address, longitude, latitude, contact, created_at, updated_at
FROM hubs
WHERE id = $1;

-- name: UpdateHub :one
UPDATE hubs
SET 
    name = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    longitude = COALESCE(sqlc.narg('longitude'), longitude),
    latitude = COALESCE(sqlc.narg('latitude'), latitude),
    contact = COALESCE(sqlc.narg('contact'), contact),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
RETURNING id, name, address, longitude, latitude, contact, created_at, updated_at;

-- name: CheckHubLinkedItems :one
SELECT COUNT(*) as item_count
FROM items
WHERE hub_id = $1 AND status != 'DELETED';

-- name: DeleteHub :exec
DELETE FROM hubs WHERE id = $1;