-- name: FetchAuditLogs :many
SELECT
    al.id,
    al.actor_id,
    al.action,
    al.target_type,
    al.target_id,
    al.created_at,
    jsonb_build_object(
        'id', u.id,
        'name', u.name,
        'email', u.email
    ) AS actor
FROM audit_logs al
LEFT JOIN users u ON u.id = al.actor_id
ORDER BY al.created_at DESC;

-- name: CreateAuditLog :exec
INSERT INTO audit_logs (actor_id, action, target_type, target_id)
VALUES ($1, $2, $3, $4);
