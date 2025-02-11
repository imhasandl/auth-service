-- +goose Up
CREATE TABLE users (
   id UUID PRIMARY KEY,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   email TEXT NOT NULL UNIQUE,
   password TEXT NOT NULL,
   username TEXT NOT NULL,
   is_premium BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE users;
