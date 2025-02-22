-- +goose Up
CREATE INDEX idx_posts_body ON posts(body);

-- +goose Down
DROP INDEX idx_posts_body;