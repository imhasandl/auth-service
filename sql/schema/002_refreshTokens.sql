-- +goose Up
CREATE TABLE refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,  -- The actual refresh token
    user_id UUID NOT NULL REFERENCES users(id), -- Foreign key to your users table
    expiry_time TIMESTAMP NOT NULL,      -- When the refresh token expires
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE refresh_tokens;