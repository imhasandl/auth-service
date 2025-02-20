-- +goose Up
CREATE TABLE reports (
   id UUID PRIMARY KEY,
   reported_at TIMESTAMP NOT NULL,
   reported_by UUID NOT NULL,
   reason TEXT NOT NULL
);

-- +goose Down
DROP TABLE reports;