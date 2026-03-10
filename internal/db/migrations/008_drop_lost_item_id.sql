-- +goose Up
-- Drop lost_item_id column from claims table
-- +goose StatementBegin
ALTER TABLE claims DROP COLUMN IF EXISTS lost_item_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims ADD COLUMN IF NOT EXISTS lost_item_id uuid REFERENCES items(id) ON UPDATE CASCADE ON DELETE RESTRICT;
-- +goose StatementEnd
