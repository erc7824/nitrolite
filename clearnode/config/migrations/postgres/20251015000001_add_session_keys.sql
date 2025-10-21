-- +goose Up
-- +goose StatementBegin

-- Create session_keys table for session keys with spending caps
CREATE TABLE session_keys (
    id SERIAL PRIMARY KEY,
    signer_address VARCHAR NOT NULL UNIQUE,
    wallet_address VARCHAR NOT NULL,
    application_name VARCHAR NOT NULL,
    allowance TEXT,
    used_allowance TEXT,
    scope VARCHAR NOT NULL DEFAULT 'all',
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
CREATE UNIQUE INDEX idx_session_keys_unique_wallet_app
  ON session_keys(wallet_address, application_name);

ALTER TABLE ledger ADD COLUMN IF NOT EXISTS session_key VARCHAR;
CREATE INDEX IF NOT EXISTS idx_ledger_session_key ON ledger(session_key);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_ledger_session_key;
ALTER TABLE ledger DROP COLUMN IF EXISTS session_key;

DROP TABLE IF EXISTS session_keys;

-- +goose StatementEnd