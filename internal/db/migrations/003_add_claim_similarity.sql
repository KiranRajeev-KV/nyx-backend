-- +goose Up
-- Add lost_item_id and similarity_score to claims table
-- +goose StatementBegin
ALTER TABLE claims ADD COLUMN IF NOT EXISTS lost_item_id uuid REFERENCES items(id) ON UPDATE CASCADE ON DELETE RESTRICT;
ALTER TABLE claims ADD COLUMN IF NOT EXISTS similarity_score double precision;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims DROP COLUMN IF EXISTS lost_item_id;
ALTER TABLE claims DROP COLUMN IF EXISTS similarity_score;
-- +goose StatementEnd
