-- +goose Up
-- Squashed migration combining all previous migrations and adding new data models

-- TODO: Review as this is a generated draft to have some foundation.

-- Channels table: Represents state channels between user and node
CREATE TABLE channels (
    channel_id VARCHAR PRIMARY KEY,
    user_wallet VARCHAR NOT NULL,
    type VARCHAR NOT NULL, -- 'escrow' or 'home'
    blockchain_id INTEGER NOT NULL,
    token VARCHAR NOT NULL,
    challenge BIGINT DEFAULT 0,
    nonce BIGINT DEFAULT 0,
    status VARCHAR NOT NULL, -- 'open', 'closed', 'challenged'
    on_chain_state_version BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- States table: Immutable state records with deterministic IDs
CREATE TABLE states (
    id VARCHAR(64) PRIMARY KEY, -- Deterministic hash: Hash(UserWallet, Asset, CycleIndex, Version)

    data TEXT,
    asset VARCHAR NOT NULL,
    user_wallet VARCHAR NOT NULL,
    cycle_index BIGINT NOT NULL,
    version BIGINT NOT NULL,

    -- Optional channel references
    home_channel_id VARCHAR,
    escrow_channel_id VARCHAR,

    -- Home Channel balances and flows
    home_user_balance DECIMAL(64,18) DEFAULT 0,
    home_user_net_flow DECIMAL(64,18) DEFAULT 0,
    home_node_balance DECIMAL(64,18) DEFAULT 0,
    home_node_net_flow DECIMAL(64,18) DEFAULT 0,

    -- Escrow Channel balances and flows
    escrow_user_balance DECIMAL(64,18) DEFAULT 0,
    escrow_user_net_flow DECIMAL(64,18) DEFAULT 0,
    escrow_node_balance DECIMAL(64,18) DEFAULT 0,
    escrow_node_net_flow DECIMAL(64,18) DEFAULT 0,

    is_final BOOLEAN DEFAULT FALSE,

    user_sig TEXT,
    node_sig TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_states_user_wallet ON states(user_wallet);
CREATE INDEX idx_states_asset ON states(asset);
CREATE INDEX idx_states_user_wallet_asset ON states(user_wallet, asset);
CREATE INDEX idx_states_home_channel_id ON states(home_channel_id) WHERE home_channel_id IS NOT NULL;
CREATE INDEX idx_states_escrow_channel_id ON states(escrow_channel_id) WHERE escrow_channel_id IS NOT NULL;

-- Ledger transactions table: Records all transactions with optional state references
CREATE TABLE ledger_transactions (
    id VARCHAR(64) PRIMARY KEY, -- Deterministic hash: Hash(To/FromAccount, Sender/ReceiverNewStateID)
    tx_type VARCHAR NOT NULL, -- 'transfer', 'commit', 'release', 'home_deposit', 'home_withdrawal', 'mutual_lock', 'escrow_deposit', 'escrow_lock', 'escrow_withdraw', 'migrate'
    asset_symbol VARCHAR NOT NULL,
    from_account VARCHAR NOT NULL,
    to_account VARCHAR NOT NULL,
    sender_new_state_id VARCHAR(64),
    receiver_new_state_id VARCHAR(64),
    amount DECIMAL(64,18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (sender_new_state_id) REFERENCES states(id) ON DELETE SET NULL,
    FOREIGN KEY (receiver_new_state_id) REFERENCES states(id) ON DELETE SET NULL
);

-- Ledger TXs are going to be used only for app sessions.
CREATE INDEX idx_ledger_transactions_type ON ledger_transactions(tx_type);
CREATE INDEX idx_ledger_transactions_from_account ON ledger_transactions(from_account);
CREATE INDEX idx_ledger_transactions_to_account ON ledger_transactions(to_account);
CREATE INDEX idx_ledger_transactions_from_to_account ON ledger_transactions(from_account, to_account, tx_type);
CREATE INDEX idx_ledger_transactions_from_account_comp ON ledger_transactions(from_account, asset_symbol, created_at DESC);
CREATE INDEX idx_ledger_transactions_to_account_comp ON ledger_transactions(to_account, asset_symbol, created_at DESC);

-- Ledger table: Balance tracking per account
CREATE TABLE ledger (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR NOT NULL,
    account_type BIGINT NOT NULL,
    asset_symbol VARCHAR NOT NULL,
    wallet VARCHAR NOT NULL,
    credit DECIMAL(64,18) NOT NULL,
    debit DECIMAL(64,18) NOT NULL,
    session_key VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_session_key ON ledger(session_key);

-- App sessions table: Application sessions
CREATE TABLE app_sessions (
    session_id VARCHAR(32) PRIMARY KEY,
    nonce BIGINT NOT NULL,
    participants TEXT[] NOT NULL,
    weights INTEGER[],
    quorum BIGINT DEFAULT 100,
    status VARCHAR NOT NULL,
    session_data TEXT NOT NULL DEFAULT '',
    application VARCHAR NOT NULL DEFAULT 'clearnode',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Session keys table: Session keys with spending caps
-- CREATE TABLE session_keys (
--     id SERIAL PRIMARY KEY,
--     address VARCHAR NOT NULL UNIQUE,
--     wallet_address VARCHAR NOT NULL,
--     application VARCHAR NOT NULL,
--     allowance JSONB,
--     scope VARCHAR NOT NULL,
--     expires_at TIMESTAMPTZ NOT NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
-- CREATE UNIQUE INDEX idx_session_keys_unique_wallet_app ON session_keys(wallet_address, application);

-- Contract events table: Blockchain event logs
CREATE TABLE contract_events (
    id BIGSERIAL PRIMARY KEY,
    contract_address VARCHAR(255) NOT NULL,
    chain_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    block_number BIGINT NOT NULL,
    transaction_hash VARCHAR(255) NOT NULL,
    log_index INTEGER NOT NULL DEFAULT 0,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX contract_events_transaction_hash_log_index_chain_idx ON contract_events (transaction_hash, log_index, chain_id);

-- Blockchain actions table: Pending blockchain operations
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

-- RPC store table: RPC request/response storage
CREATE TABLE rpc_store (
    id SERIAL PRIMARY KEY,
    req_id BIGINT NOT NULL,
    msg_type INT NOT NULL, -- 1 for request, 2 for response, 3 for event
    method VARCHAR(255) NOT NULL,
    payload TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
);

-- +goose Down
DROP TABLE IF EXISTS rpc_store;
DROP INDEX IF EXISTS idx_blockchain_actions_pending;
DROP TABLE IF EXISTS blockchain_actions;
DROP INDEX IF EXISTS contract_events_transaction_hash_log_index_chain_idx;
DROP TABLE IF EXISTS contract_events;
DROP INDEX IF EXISTS idx_session_keys_unique_wallet_app;
DROP INDEX IF EXISTS idx_session_keys_wallet_address;
DROP TABLE IF EXISTS session_keys;
DROP TABLE IF EXISTS app_sessions;
DROP INDEX IF EXISTS idx_ledger_session_key;
DROP TABLE IF EXISTS ledger;
DROP INDEX IF EXISTS idx_ledger_transactions_to_account_comp;
DROP INDEX IF EXISTS idx_ledger_transactions_from_account_comp;
DROP INDEX IF EXISTS idx_ledger_transactions_from_to_account;
DROP INDEX IF EXISTS idx_ledger_transactions_to_account;
DROP INDEX IF EXISTS idx_ledger_transactions_from_account;
DROP INDEX IF EXISTS idx_ledger_transactions_type;
DROP TABLE IF EXISTS ledger_transactions;
DROP INDEX IF EXISTS idx_states_escrow_channel_id;
DROP INDEX IF EXISTS idx_states_home_channel_id;
DROP INDEX IF EXISTS idx_states_user_wallet_asset;
DROP INDEX IF EXISTS idx_states_asset;
DROP INDEX IF EXISTS idx_states_user_wallet;
DROP TABLE IF EXISTS states;
DROP TABLE IF EXISTS channels;
