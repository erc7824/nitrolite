-- +goose Up
-- +goose StatementBegin

ALTER TABLE app_sessions ADD COLUMN session_data TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE app_sessions DROP COLUMN session_data;

-- +goose StatementEnd
