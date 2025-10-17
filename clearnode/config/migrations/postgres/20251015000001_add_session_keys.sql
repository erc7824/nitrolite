-- +goose Up
-- +goose StatementBegin

-- Create the new unified session_keys table
CREATE TABLE session_keys (
    id SERIAL PRIMARY KEY,
    signer_address VARCHAR NOT NULL,
    wallet_address VARCHAR NOT NULL,
    application_name VARCHAR,
    spending_cap TEXT,
    used_allowance TEXT,
    scope VARCHAR,
    expiration_time TIMESTAMP,
    signer_type VARCHAR NOT NULL DEFAULT 'session',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add indexes for performance
CREATE INDEX idx_session_keys_signer_address ON session_keys(signer_address);
CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
CREATE INDEX idx_session_keys_signer_type ON session_keys(signer_type);

-- Migrate existing data from signers table if it exists
INSERT INTO session_keys (signer_address, wallet_address, signer_type, created_at, updated_at)
SELECT 
    signer as signer_address,
    wallet as wallet_address,
    'custody' as signer_type,
    CURRENT_TIMESTAMP as created_at,
    CURRENT_TIMESTAMP as updated_at
FROM signers
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'signers');

-- Drop the old signers table if it exists
DROP TABLE IF EXISTS signers;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate the old signers table
CREATE TABLE signers (
    signer VARCHAR PRIMARY KEY,
    wallet VARCHAR NOT NULL
);

-- Migrate custody signers back to signers table
INSERT INTO signers (signer, wallet)
SELECT signer_address, wallet_address 
FROM session_keys 
WHERE signer_type = 'custody';

-- Drop the unified session_keys table
DROP TABLE session_keys;

-- +goose StatementEnd