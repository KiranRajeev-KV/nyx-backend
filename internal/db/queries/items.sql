-- name: FetchAllItems :many
SELECT
    id,
    name,
    description,
    image_url_redacted,
    status,
    type,
    created_at,
    updated_at
FROM
    items
WHERE
    status = 'OPEN'
    OR status = 'PENDING_CLAIM'
ORDER BY
    created_at DESC;

-- name: FetchItemsByType :many
SELECT
    id,
    name,
    description,
    image_url_redacted,
    status,
    type,
    created_at,
    updated_at
FROM
    items
WHERE
    type = $1
    AND (status = 'OPEN'
        OR status = 'PENDING_CLAIM')
ORDER BY
    created_at DESC;

-- name: CreateItem :one
INSERT INTO items (
    user_id,
    is_anonymous,
    hub_id,
    name,
    description,
    type,
    location_description,
    time_at,
    latitude,
    longitude,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW()
)
RETURNING
    id,
    user_id,
    is_anonymous,
    hub_id,
    name,
    description,
    type,
    location_description,
    time_at,
    latitude,
    longitude,
    created_at,
    updated_at
;
