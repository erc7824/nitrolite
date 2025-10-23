-- +goose Up
-- +goose StatementBegin

DROP TABLE IF EXISTS signers;

-- Create session_keys table for session keys with spending caps
CREATE TABLE session_keys (
    id SERIAL PRIMARY KEY,
    signer_address VARCHAR NOT NULL UNIQUE,
    wallet_address VARCHAR NOT NULL,
    app_name VARCHAR NOT NULL,
    app_address VARCHAR NOT NULL DEFAULT '',
    allowance TEXT,
    used_allowance TEXT,
    scope VARCHAR NOT NULL DEFAULT 'all',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
-- Ensure one session key per wallet+app (identified by both name and address together)
CREATE UNIQUE INDEX idx_session_keys_unique_wallet_app
  ON session_keys(wallet_address, app_name);

ALTER TABLE ledger ADD COLUMN IF NOT EXISTS session_key VARCHAR;
CREATE INDEX IF NOT EXISTS idx_ledger_session_key ON ledger(session_key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_ledger_session_key;
ALTER TABLE ledger DROP COLUMN IF EXISTS session_key;

DROP TABLE IF EXISTS session_keys;

CREATE TABLE signers (
    signer VARCHAR PRIMARY KEY,
    wallet VARCHAR NOT NULL
);

-- +goose StatementEnd
