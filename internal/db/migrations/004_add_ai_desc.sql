-- +goose Up
-- Add ai_desc column to items table and update search_text to include it
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN IF NOT EXISTS ai_desc text;

-- Update search_text to include ai_desc
ALTER TABLE items DROP COLUMN IF EXISTS search_text;
ALTER TABLE items ADD COLUMN IF NOT EXISTS search_text tsvector
    GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || COALESCE(description, '') || ' ' || COALESCE(ai_desc, ''))) STORED;

CREATE INDEX IF NOT EXISTS idx_items_search ON items USING GIN(search_text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_search;
ALTER TABLE items DROP COLUMN IF EXISTS search_text;
ALTER TABLE items DROP COLUMN IF EXISTS ai_desc;

-- Restore original search_text without ai_desc
ALTER TABLE items ADD COLUMN IF NOT EXISTS search_text tsvector
    GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || COALESCE(description, ''))) STORED;
CREATE INDEX IF NOT EXISTS idx_items_search ON items USING GIN(search_text);
-- +goose StatementEnd
