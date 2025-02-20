-- +goose Up
ALTER TABLE posts ADD COLUMN liked_by TEXT[];

-- +goose Down
ALTER TABLE posts DROP COLUMN liked_by;