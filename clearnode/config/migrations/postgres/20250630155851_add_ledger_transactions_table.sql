-- +goose Up
-- +goose StatementBegin
CREATE TABLE ledger_transactions (
    id SERIAL PRIMARY KEY,
    tx_type INTEGER NOT NULL,
    from_account VARCHAR NOT NULL,
    to_account VARCHAR NOT NULL,
    asset_symbol VARCHAR NOT NULL,
    amount DECIMAL(64,18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_transactions_from_account_comp ON ledger_transactions(from_account, asset_symbol, created_at DESC);
CREATE INDEX idx_ledger_transactions_to_account_comp ON ledger_transactions(to_account, asset_symbol, created_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_ledger_transactions_to_account_comp;
DROP INDEX IF EXISTS idx_ledger_transactions_from_account_comp;

DROP TABLE IF EXISTS ledger_transactions;
-- +goose StatementEnd
