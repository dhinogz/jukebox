-- +goose Up
-- +goose StatementBegin
CREATE TABLE oauth (
    id SERIAL PRIMARY KEY,
    user_id INT UNIQUE REFERENCES users (id) ON DELETE CASCADE,
    provider TEXT,
    access_token TEXT,
    refresh_token TEXT,
    token_type TEXT,
    expiry timestamp with time zone
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE oauth;
-- +goose StatementEnd

-- Add not null and unique index on (user_id, service)
