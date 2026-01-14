-- +goose Up
-- Extensions
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "vector";

-- +goose StatementEnd
-- Enums
-- +goose StatementBegin
CREATE TYPE user_role AS ENUM(
    'USER',
    'ADMIN'
);

CREATE TYPE item_status AS ENUM(
    'OPEN',
    'PENDING_CLAIM',
    'RESOLVED',
    'ARCHIVED',
    'DELETED'
);

CREATE TYPE item_type AS ENUM(
    'FOUND',
    'LOST'
);

CREATE TYPE claim_status AS ENUM(
    'PENDING',
    'APPROVED',
    'REJECTED'
);

CREATE TYPE target_type AS ENUM(
    'ITEM',
    'USER',
    'HUB',
    'CLAIM'
);

-- +goose StatementEnd
-- === Tables ===
-- 1. Users
-- +goose StatementBegin
CREATE TABLE users(
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    phone text,
    password TEXT NOT NULL,
    ROLE user_role NOT NULL DEFAULT 'USER',
    refresh_token text,
    trust_score integer DEFAULT 100,
    is_verified boolean NOT NULL DEFAULT FALSE,
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW()
);

-- +goose StatementEnd
-- 2.UserOnboarding
-- +goose StatementBegin
CREATE TABLE user_onboarding(
    id serial PRIMARY KEY,
    name text NOT NULL,
    email text UNIQUE NOT NULL,
    password TEXT NOT NULL,
    otp text NOT NULL,
    attempts integer DEFAULT 0,
    verified_at timestamptz,
    created_at timestamptz DEFAULT NOW(),
    expires_at timestamptz NOT NULL
);

-- +goose StatementEnd
-- 3. Hubs
-- +goose StatementBegin
CREATE TABLE hubs(
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    name text NOT NULL,
    address text,
    longitude text,
    latitude text,
    contact text,
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW()
);

-- +goose StatementEnd
-- 4. Items
-- +goose StatementBegin
CREATE TABLE items(
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    user_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    is_anonymous boolean NOT NULL DEFAULT FALSE,
    hub_id uuid REFERENCES hubs(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    name text NOT NULL,
    image_url_original text,
    image_url_redacted text,
    embedding VECTOR(512),
    description text,
    status item_status NOT NULL DEFAULT 'OPEN',
    tags jsonb DEFAULT '[]'::jsonb,
    type item_type NOT NULL,
    location_description text,
    time_at timestamptz,
    latitude text,
    longitude text,
    metadata jsonb DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW()
);

-- +goose StatementEnd
-- 5. Claims
-- +goose StatementBegin
CREATE TABLE claims(
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    item_id uuid NOT NULL REFERENCES items(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    claimant_id uuid NOT NULL REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    status claim_status NOT NULL DEFAULT 'PENDING',
    similarity_score float,
    proof_text text,
    proof_image_url text,
    admin_notes text,
    processed_by uuid REFERENCES users(id) ON UPDATE CASCADE ON DELETE RESTRICT,
    created_at timestamptz DEFAULT NOW(),
    updated_at timestamptz DEFAULT NOW()
);

-- +goose StatementEnd
-- 6. AuditLogs
-- +goose StatementBegin
CREATE TABLE audit_logs(
    id uuid PRIMARY KEY DEFAULT uuidv7(),
    actor_id uuid REFERENCES users(id) ON UPDATE CASCADE,
    action text NOT NULL,
    target_type target_type NOT NULL,
    target_id uuid,
    created_at timestamptz DEFAULT NOW()
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

DROP TABLE IF EXISTS user_onboarding;

DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS claim_status;

DROP TYPE IF EXISTS item_type;

DROP TYPE IF EXISTS item_status;

DROP TYPE IF EXISTS user_role;

DROP TYPE IF EXISTS target_type;

DROP EXTENSION IF EXISTS "vector";

-- +goose StatementEnd
