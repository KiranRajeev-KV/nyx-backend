-- +goose Up
-- Extensions
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "vector";
-- +goose StatementEnd

-- Enums
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM ('USER', 'ADMIN');
CREATE TYPE item_status AS ENUM ('OPEN', 'PENDING_CLAIM', 'RESOLVED', 'ARCHIVED', 'DELETED');
CREATE TYPE item_type AS ENUM ('FOUND', 'LOST');
CREATE TYPE claim_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED');
CREATE TYPE target_type AS ENUM ('ITEM', 'USER', 'HUB', 'CLAIM');
-- +goose StatementEnd

-- === Tables ===

-- 1. Users
-- +goose StatementBegin
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    password TEXT NOT NULL,
    role user_role NOT NULL,
    refresh_token TEXT,
    trust_score INTEGER DEFAULT 100,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- 2. Hubs
-- +goose StatementBegin
CREATE TABLE hubs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    name TEXT NOT NULL,
    address TEXT,
    longitude TEXT,
    latitude TEXT,
    contact TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- 3. Items
-- +goose StatementBegin
CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    is_anonymous BOOLEAN DEFAULT FALSE,
    hub_id UUID REFERENCES hubs(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name TEXT NOT NULL,
    image_url_original TEXT,
    image_url_redacted TEXT,
    embedding VECTOR(512),
    description TEXT,
    status item_status NOT NULL DEFAULT 'OPEN',
    tags JSONB DEFAULT '[]'::JSONB,
    type item_type NOT NULL,

    location_description TEXT,
    time_at TIMESTAMPTZ,
    latitude TEXT,
    longitude TEXT,

    metadata JSONB DEFAULT '{}'::JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- 4. Claims
-- +goose StatementBegin
CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    item_id UUID NOT NULL REFERENCES items(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    claimant_id UUID NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    status claim_status NOT NULL DEFAULT 'PENDING',
    similarity_score FLOAT,
    proof_text TEXT,
    proof_image_url TEXT,
    admin_notes TEXT,
    processed_by UUID REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- 5. AuditLogs
-- +goose StatementBegin
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    actor_id UUID REFERENCES users(id) ON UPDATE CASCADE,
    action TEXT NOT NULL,
    target_type target_type NOT NULL,
    target_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- Indexes
-- +goose StatementBegin
CREATE INDEX idx_items_type ON items(type);

CREATE INDEX idx_items_user_id ON items(user_id);
CREATE INDEX idx_items_hub_id ON items(hub_id);

-- TODO: Add vector indexing
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_type;
DROP INDEX IF EXISTS idx_items_user_id;
DROP INDEX IF EXISTS idx_items_hub_id;

DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS items;
DROP TABLE IF EXISTS hubs;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS claim_status;
DROP TYPE IF EXISTS item_type;
DROP TYPE IF EXISTS item_status;
DROP TYPE IF EXISTS user_role;
DROP TYPE IF EXISTS target_type;

DROP EXTENSION IF EXISTS "vector";
-- +goose StatementEnd