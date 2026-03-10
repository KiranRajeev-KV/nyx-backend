-- +goose Up
-- Drop image_url_redacted column, use image_url_original instead
-- +goose StatementBegin
ALTER TABLE items DROP COLUMN IF EXISTS image_url_redacted;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE items ADD COLUMN IF NOT EXISTS image_url_redacted text;
-- +goose StatementEnd
