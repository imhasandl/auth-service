-- +goose Up
CREATE TABLE posts (
   id UUID PRIMARY KEY,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   posted_by TEXT NOT NULL,
   body TEXT NOT NULL,
   likes INT NOT NULL DEFAULT 0,
   views INT NOT NULL DEFAULT 0,
   liked_by TEXT[]
);

-- +goose Down
DROP TABLE posts;