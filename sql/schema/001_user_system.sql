-- +goose Up
CREATE TABLE users (
   id UUID NOT NULL PRIMARY KEY,
   created_at TIMESTAMP NOT NULL,
   updated_at TIMESTAMP NOT NULL,
   email TEXT NOT NULL UNIQUE,
   password TEXT NOT NULL,
   username TEXT NOT NULL,
   subscribers UUID[],
   subscribed_to UUID[],
   is_premium BOOLEAN NOT NULL DEFAULT FALSE,
   verification_code INT NOT NULL,
   verification_expire_time TIMESTAMP NOT NULL,
   is_verified BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_users_username ON users(username);

CREATE TABLE device_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_token TEXT NOT NULL,
    device_type TEXT NOT NULL, -- e.g., 'android', 'ios', 'web'
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_token)
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);

CREATE TABLE refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    expiry_time TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE users;
DROP INDEX idx_users_username;
DROP TABLE device_token;
DROP INDEX idx_device_tokens_user_id;
DROP TABLE refresh_tokens;
