-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_tags (
    wallet VARCHAR(42) PRIMARY KEY,
    tag VARCHAR(20) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_tags;
-- +goose StatementEnd
