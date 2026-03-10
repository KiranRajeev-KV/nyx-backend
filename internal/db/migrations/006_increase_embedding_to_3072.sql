-- +goose Up
-- Increase embedding dimensions from 768 to 3072 to match default Gemini embedding output
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_embedding;
ALTER TABLE items ALTER COLUMN embedding TYPE VECTOR(3072);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_items_embedding;
ALTER TABLE items ALTER COLUMN embedding TYPE VECTOR(768);
-- +goose StatementEnd
