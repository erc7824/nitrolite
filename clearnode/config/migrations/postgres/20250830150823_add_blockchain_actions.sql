-- +goose Up
CREATE TABLE blockchain_actions (
    id BIGSERIAL PRIMARY KEY,
    action_type VARCHAR(50) NOT NULL,
    channel_id VARCHAR(66) NOT NULL,
    chain_id INTEGER NOT NULL,
    action_data JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    transaction_hash VARCHAR(66),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_blockchain_actions_channel
        FOREIGN KEY(channel_id)
        REFERENCES channels(channel_id)
        ON DELETE CASCADE
);


CREATE INDEX idx_blockchain_actions_pending ON blockchain_actions(status, created_at) WHERE status = 'pending';

-- +goose Down
DROP INDEX idx_blockchain_actions_pending;
DROP TABLE blockchain_actions;
