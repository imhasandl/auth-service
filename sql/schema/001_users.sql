-- +goose Up
CREATE TABLE users (
   id UUID NOT NULL PRIMARY KEY,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   email TEXT NOT NULL UNIQUE,
   password TEXT NOT NULL,
   username TEXT NOT NULL,
   is_premium BOOLEAN NOT NULL DEFAULT FALSE,
   verification_code INT NOT NULL,
   is_verified BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE users;
