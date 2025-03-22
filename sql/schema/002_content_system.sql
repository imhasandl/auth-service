-- +goose Up
CREATE TABLE posts (
   id UUID PRIMARY KEY,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   posted_by UUID NOT NULL,
   body TEXT NOT NULL,
   likes INT NOT NULL DEFAULT 0,
   views INT NOT NULL DEFAULT 0,
   liked_by TEXT[]
);

CREATE INDEX idx_posts_body ON posts(body);

CREATE TABLE comments (
   id UUID NOT NULL PRIMARY KEY, 
   created_at TIMESTAMP NOT NULL, 
   post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
   user_id UUID NOT NULL REFERENCES users(id),
   comment_text TEXT NOT NULL
);

-- +goose Down
DROP TABLE comments;
DROP INDEX idx_posts_body;
DROP TABLE posts;