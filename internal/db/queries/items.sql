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

-- name: FetchItemByID :one
SELECT
    i.id,
    i.user_id,
    i.is_anonymous,
    i.hub_id,
    i.name,
    i.image_url_redacted,
    i.description,
    i.status,
    i.type,
    i.location_description,
    i.time_at,
    i.latitude,
    i.longitude,
    i.created_at,
    i.updated_at,

    -- user details, null if found & anonymous is false
    CASE
      WHEN i.type = 'FOUND' AND i.is_anonymous THEN NULL
      ELSE jsonb_build_object(
        'id', u.id,
        'name', u.name,
        'email', u.email,
        'phone', u.phone,
        'trust_score', u.trust_score
      )
    END AS "user",

    -- hub details only for FOUND items
    CASE
      WHEN i.type = 'FOUND' THEN jsonb_build_object(
        'id', h.id,
        'name', h.name,
        'address', h.address,
        'contact', h.contact
      )
      ELSE NULL
    END AS hub

FROM items i
LEFT JOIN users u ON u.id = i.user_id
LEFT JOIN hubs h ON h.id = i.hub_id
WHERE i.id = $1;

-- name: FetchAllItemsByUserId :many
SELECT
    i.id,
    i.user_id,
    i.is_anonymous,
    i.hub_id,
    i.name,
    i.image_url_redacted,
    i.description,
    i.status,
    i.type,
    i.location_description,
    i.time_at,
    i.latitude,
    i.longitude,
    i.created_at,
    i.updated_at,
    
    jsonb_build_object(
        'id', u.id,
        'name', u.name,
        'email', u.email,
        'phone', u.phone,
        'trust_score', u.trust_score
    ) AS "user",

    -- hub details only for FOUND items
    CASE
      WHEN i.type = 'FOUND' THEN jsonb_build_object(
        'id', h.id,
        'name', h.name,
        'address', h.address,
        'contact', h.contact
      )
      ELSE NULL
    END AS hub

FROM items i
LEFT JOIN users u ON u.id = i.user_id
LEFT JOIN hubs h ON h.id = i.hub_id
WHERE i.user_id = $1 AND i.status != 'DELETED'
ORDER BY i.created_at DESC;

-- name: UpdateItemById :one
UPDATE items
SET
    name = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    location_description = COALESCE(sqlc.narg('location_description'), location_description),
    time_at = COALESCE(sqlc.narg('time_at'), time_at),
    latitude = COALESCE(sqlc.narg('latitude'), latitude),
    longitude = COALESCE(sqlc.narg('longitude'), longitude),
    hub_id = COALESCE(sqlc.narg('hub_id'), hub_id),
    updated_at = NOW()
WHERE id = sqlc.arg('id')
  AND user_id = sqlc.arg('user_id')
RETURNING
    id,
    hub_id,
    name,
    description,
    location_description,
    time_at,
    latitude,
    longitude,
    updated_at;

-- name: SoftDeleteItemById :one
UPDATE
    items
SET
    status = 'DELETED',
    updated_at = NOW()
WHERE
    id = $1
    AND user_id = $2
RETURNING
    id,
    status,
    updated_at;

-- name: UpdateItemStatusById :one
UPDATE
    items
SET
    status = $2,
    updated_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    status,
    updated_at;