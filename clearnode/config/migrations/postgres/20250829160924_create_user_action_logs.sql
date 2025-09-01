-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS user_action_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    label VARCHAR(255) NOT NULL,
    metadata TEXT,
    created_at TIMESTAMPTZ NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_action_logs_user_id ON user_action_logs (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_user_action_logs_user_id;
DROP TABLE IF EXISTS user_action_logs;
-- +goose StatementEnd