-- +goose Up 
CREATE TABLE comments (
   id UUID NOT NULL PRIMARY KEY, 
   created_at TIMESTAMP NOT NULL, 
   post_id UUID NOT NULL REFERENCES posts(id),
   user_id UUID NOT NULL REFERENCES users(id),
   comment_text TEXT NOT NULL
);

-- +goose Down
DROP TABLE comments;
