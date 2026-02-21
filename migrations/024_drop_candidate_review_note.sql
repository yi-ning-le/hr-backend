-- +goose Up
ALTER TABLE candidates DROP COLUMN review_note;

-- +goose Down
ALTER TABLE candidates ADD COLUMN review_note TEXT;
