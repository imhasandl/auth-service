-- +goose Up
CREATE TABLE messages (
    id UUID PRIMARY KEY,
    sent_at TIMESTAMP NOT NULL,
    sender_id UUID REFERENCES users(id) NOT NULL,
    receiver_id UUID REFERENCES users(id) NOT NULL,
    content TEXT NOT NULL
);

CREATE TABLE reports (
   id UUID PRIMARY KEY,
   reported_at TIMESTAMP NOT NULL,
   reported_by UUID NOT NULL,
   reason TEXT NOT NULL
);

-- +goose Down
DROP TABLE reports;
DROP TABLE messages;