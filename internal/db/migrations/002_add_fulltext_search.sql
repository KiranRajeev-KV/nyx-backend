-- +goose Up
-- Add full-text search for items
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN IF NOT EXISTS search_text tsvector
    GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || COALESCE(description, ''))) STORED;

CREATE INDEX IF NOT EXISTS idx_items_search ON items USING GIN(search_text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_search;
ALTER TABLE items DROP COLUMN IF EXISTS search_text;
-- +goose StatementEnd
