-- +goose Up
-- Increase embedding dimensions from 512 to 768 to match Gemini embedding output
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_embedding;
ALTER TABLE items ALTER COLUMN embedding TYPE VECTOR(768);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_embedding;
ALTER TABLE items ALTER COLUMN embedding TYPE VECTOR(512);
-- +goose StatementEnd
