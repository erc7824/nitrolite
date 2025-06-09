-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS contract_events (
                                               id BIGSERIAL PRIMARY KEY,
                                               contract_address VARCHAR(255) NOT NULL,
                                               chain_id BIGINT NOT NULL,
                                               name VARCHAR(255) NOT NULL,
                                               block_number BIGINT NOT NULL,
                                               transaction_hash VARCHAR(255) NOT NULL,
                                               log_index INTEGER NOT NULL DEFAULT 0,
                                               data JSONB NOT NULL,
                                               created_at TIMESTAMPTZ NULL DEFAULT now()
);

CREATE UNIQUE INDEX contract_events_transaction_hash_log_index_chain_idx ON contract_events (transaction_hash, log_index, chain_id);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS contract_events_transaction_hash_log_index_chain_idx;
DROP TABLE IF EXISTS contract_events;

-- +goose StatementEnd
