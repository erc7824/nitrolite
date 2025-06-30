-- +goose Up
-- +goose StatementBegin
CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    hash VARCHAR NOT NULL UNIQUE,
    tx_type INTEGER NOT NULL,
    from_account VARCHAR NOT NULL,
    to_account VARCHAR NOT NULL,
    asset_symbol VARCHAR NOT NULL,
    amount DECIMAL(64,18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create indexes for optimal query performance
CREATE INDEX idx_transactions_type ON transactions(tx_type);
CREATE INDEX idx_transactions_from_account ON transactions(from_account);
CREATE INDEX idx_transactions_to_account ON transactions(to_account);
CREATE INDEX idx_transactions_from_to_account ON transactions(from_account, to_account);
CREATE INDEX idx_transactions_created_at ON transactions(created_at DESC);
CREATE INDEX idx_transactions_asset_symbol ON transactions(asset_symbol);

-- Composite indexes for common query patterns
CREATE INDEX idx_transactions_from_account_asset ON transactions(from_account, asset_symbol);
CREATE INDEX idx_transactions_to_account_asset ON transactions(to_account, asset_symbol);
CREATE INDEX idx_transactions_type_asset ON transactions(tx_type, asset_symbol);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE transactions;
-- +goose StatementEnd
