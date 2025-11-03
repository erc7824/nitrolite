-- +goose Up
-- +goose StatementBegin

-- NOTE: We would keep the old signers table in case we want to revert

-- Create session_keys table for session keys with spending caps
CREATE TABLE session_keys (
    id SERIAL PRIMARY KEY,
    address VARCHAR NOT NULL UNIQUE,
    wallet_address VARCHAR NOT NULL,
    application VARCHAR NOT NULL,
    allowance JSONB,
    scope VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
-- Ensure one session key per wallet+app (identified by both name and address together)
CREATE UNIQUE INDEX idx_session_keys_unique_wallet_app
  ON session_keys(wallet_address, application);

ALTER TABLE ledger ADD COLUMN IF NOT EXISTS session_key VARCHAR;
CREATE INDEX IF NOT EXISTS idx_ledger_session_key ON ledger(session_key);

ALTER TABLE app_sessions ADD COLUMN IF NOT EXISTS application VARCHAR NOT NULL DEFAULT 'clearnode';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE app_sessions DROP COLUMN IF EXISTS application;

DROP INDEX IF EXISTS idx_ledger_session_key;
ALTER TABLE ledger DROP COLUMN IF EXISTS session_key;

DROP INDEX IF EXISTS idx_session_keys_unique_wallet_app;
DROP INDEX IF EXISTS idx_session_keys_wallet_address;
DROP TABLE IF EXISTS session_keys;

-- +goose StatementEnd
