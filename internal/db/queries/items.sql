-- name: FetchAllItems :many
SELECT
    id,
    name,
    description,
    image_url_redacted,
    status,
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

