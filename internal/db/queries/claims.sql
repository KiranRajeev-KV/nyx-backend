-- name: CreateClaim :one
INSERT INTO claims (
    item_id, claimant_id, proof_text, proof_image_url, similarity_score, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, NOW(), NOW()
)
RETURNING 
    id, item_id, claimant_id, status, similarity_score, created_at, updated_at;

-- name: FetchClaimsByUser :many
SELECT 
    c.id,
    c.item_id,
    c.claimant_id,
    c.status,
    c.proof_text,
    c.proof_image_url,
    c.admin_notes,
    c.created_at,
    c.updated_at,
    i.name as item_name,
    i.type as item_type,
    i.status as item_status
FROM claims c
JOIN items i ON c.item_id = i.id
WHERE c.claimant_id = $1
ORDER BY c.created_at DESC;

-- name: FetchClaimsByItem :many
SELECT 
    c.id,
    c.item_id,
    c.claimant_id,
    c.status,
    c.proof_text,
    c.proof_image_url,
    c.admin_notes,
    c.created_at,
    c.updated_at,
    u.name as claimant_name,
    u.email as claimant_email
FROM claims c
JOIN users u ON c.claimant_id = u.id
WHERE c.item_id = $1
ORDER BY c.created_at DESC;

-- name: FetchAllClaims :many
SELECT 
    c.id,
    c.item_id,
    c.claimant_id,
    c.status,
    c.proof_text,
    c.proof_image_url,
    c.admin_notes,
    c.created_at,
    c.updated_at,
    i.name as item_name,
    i.type as item_type,
    i.status as item_status,
    u.name as claimant_name,
    u.email as claimant_email
FROM claims c
JOIN items i ON c.item_id = i.id
JOIN users u ON c.claimant_id = u.id
ORDER BY c.created_at DESC;

-- name: ProcessClaim :one
UPDATE claims 
SET 
    status = $2,
    admin_notes = $3,
    processed_by = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING 
    id, status, admin_notes, processed_by, updated_at;

-- name: FetchClaimByID :one
SELECT 
    c.id,
    c.item_id,
    c.claimant_id,
    c.status,
    c.proof_text,
    c.proof_image_url,
    c.admin_notes,
    c.created_at,
    c.updated_at,
    i.name as item_name,
    i.type as item_type,
    i.status as item_status,
    u.name as claimant_name,
    u.email as claimant_email
FROM claims c
JOIN items i ON c.item_id = i.id
JOIN users u ON c.claimant_id = u.id
WHERE c.id = $1;

-- name: CheckExistingClaim :one
SELECT id, status
FROM claims 
WHERE item_id = $1 AND claimant_id = $2
LIMIT 1;

-- name: GetPendingClaimsCount :one
SELECT COUNT(*) as count
FROM claims 
WHERE item_id = $1 AND status = 'PENDING';

-- name: GetItemByID :one
SELECT id, user_id, type, status
FROM items 
WHERE id = $1;

-- name: GetItemEmbedding :one
SELECT embedding::float4[] as embedding FROM items WHERE id = $1;

-- name: UpdateItemStatus :one
UPDATE items 
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, status, updated_at;

-- name: UpdateClaimProofImage :one
UPDATE claims
SET
    proof_image_url = $2,
    updated_at = NOW()
WHERE id = $1 AND claimant_id = $3
RETURNING id, proof_image_url;