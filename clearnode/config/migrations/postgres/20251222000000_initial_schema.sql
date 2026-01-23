-- +goose Up
-- Squashed migration combining all previous migrations and adding new data models

-- Channels table: Represents state channels between user and node
CREATE TABLE channels (
    channel_id CHAR(66) PRIMARY KEY,
    user_wallet CHAR(42) NOT NULL,
    type SMALLINT NOT NULL, -- ChannelType enum: 0=void, 1=home, 2=escrow
    blockchain_id INTEGER NOT NULL, -- uint32
    token CHAR(42) NOT NULL,
    challenge_duration BIGINT NOT NULL DEFAULT 0,
    challenge_expires_at TIMESTAMPTZ,
    nonce BIGINT NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL, -- ChannelStatus enum: 0=void, 1=open, 2=challenged, 3=closed
    state_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_user_wallet ON channels(user_wallet);
CREATE INDEX idx_channels_status ON channels(status);

-- Channel States table: Immutable state records
CREATE TABLE channel_states (
    id CHAR(66) PRIMARY KEY, -- Deterministic hash: Hash(UserWallet, Asset, Epoch, Version)
    transitions JSONB NOT NULL, -- JSON array of state transitions
    asset VARCHAR(20) NOT NULL,
    user_wallet CHAR(42) NOT NULL,
    epoch BIGINT NOT NULL,
    version BIGINT NOT NULL,

    -- Optional channel references
    home_channel_id CHAR(66),
    escrow_channel_id CHAR(66),

    -- Home Channel balances and flows (balances are positive only, net flows can be negative)
    home_user_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_user_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_node_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_node_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,

    -- Escrow Channel balances and flows (balances are positive only, net flows can be negative)
    escrow_user_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_user_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_node_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_node_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,

    user_sig TEXT, -- TODO: consider using fixed char length
    node_sig TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channel_states_user_wallet ON channel_states(user_wallet);
CREATE INDEX idx_channel_states_asset ON channel_states(asset);
CREATE INDEX idx_channel_states_user_wallet_asset ON channel_states(user_wallet, asset);
CREATE INDEX idx_channel_states_epoch_version ON channel_states(epoch DESC, version DESC);
CREATE INDEX idx_channel_states_home_channel_id ON channel_states(home_channel_id) WHERE home_channel_id IS NOT NULL;
CREATE INDEX idx_channel_states_escrow_channel_id ON channel_states(escrow_channel_id) WHERE escrow_channel_id IS NOT NULL;

-- Transactions table: Records all transactions with optional state references
CREATE TABLE transactions (
    id CHAR(66) PRIMARY KEY, -- Deterministic hash
    tx_type SMALLINT NOT NULL, -- TransactionType enum
    asset_symbol VARCHAR(20) NOT NULL,
    from_account VARCHAR(66) NOT NULL, -- Can be wallet (42) or channel ID (66)
    to_account VARCHAR(66) NOT NULL, -- Can be wallet (42) or channel ID (66)
    sender_new_state_id CHAR(66),
    receiver_new_state_id CHAR(66),
    amount NUMERIC(78, 18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    FOREIGN KEY (sender_new_state_id) REFERENCES channel_states(id) ON DELETE SET NULL,
    FOREIGN KEY (receiver_new_state_id) REFERENCES channel_states(id) ON DELETE SET NULL
);

CREATE INDEX idx_transactions_type ON transactions(tx_type);
CREATE INDEX idx_transactions_from_account ON transactions(from_account);
CREATE INDEX idx_transactions_to_account ON transactions(to_account);
CREATE INDEX idx_transactions_from_to_type ON transactions(from_account, to_account, tx_type);
CREATE INDEX idx_transactions_from_comp ON transactions(from_account, asset_symbol, created_at DESC);
CREATE INDEX idx_transactions_to_comp ON transactions(to_account, asset_symbol, created_at DESC);

-- App Sessions table: Application sessions
CREATE TABLE app_sessions_v1 (
    id CHAR(66) PRIMARY KEY,
    application VARCHAR NOT NULL,
    nonce BIGINT NOT NULL,
    session_data TEXT NOT NULL,
    quorum SMALLINT NOT NULL DEFAULT 100,
    version BIGINT NOT NULL DEFAULT 1,
    status SMALLINT NOT NULL, -- AppSessionStatus enum
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_sessions_v1_application ON app_sessions_v1(application);
CREATE INDEX idx_app_sessions_v1_status ON app_sessions_v1(status);

-- App Session Participants table: Participants in application sessions
CREATE TABLE app_session_participants_v1 (
    app_session_id CHAR(66) NOT NULL,
    wallet_address CHAR(42) NOT NULL,
    signature_weight SMALLINT NOT NULL,
    PRIMARY KEY (app_session_id, wallet_address),
    FOREIGN KEY (app_session_id) REFERENCES app_sessions_v1(id) ON DELETE CASCADE
);

CREATE INDEX idx_app_session_participants_v1_wallet ON app_session_participants_v1(wallet_address);

-- App Ledger table: Internal ledger entries for application sessions
CREATE TABLE app_ledger_v1 (
    id CHAR(36) PRIMARY KEY, -- UUID
    account_id CHAR(66) NOT NULL, -- Session ID
    asset_symbol VARCHAR(20) NOT NULL,
    wallet CHAR(42) NOT NULL,
    credit NUMERIC(78, 18) NOT NULL DEFAULT 0,
    debit NUMERIC(78, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_ledger_v1_account_asset ON app_ledger_v1(account_id, asset_symbol);
CREATE INDEX idx_app_ledger_v1_wallet ON app_ledger_v1(wallet);

-- Contract events table: Blockchain event logs
CREATE TABLE contract_events (
    id BIGSERIAL PRIMARY KEY,
    contract_address VARCHAR(255) NOT NULL,
    chain_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    block_number BIGINT NOT NULL,
    transaction_hash VARCHAR(255) NOT NULL,
    log_index INTEGER NOT NULL DEFAULT 0,
    data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX contract_events_tx_log_chain_idx ON contract_events (transaction_hash, log_index, chain_id);
CREATE INDEX idx_contract_events_block ON contract_events(chain_id, block_number);

-- Blockchain actions table: Pending blockchain operations
CREATE TABLE blockchain_actions (
    id BIGSERIAL PRIMARY KEY,
    action_type SMALLINT NOT NULL,
    state_id CHAR(66),
    action_data JSONB,
    status SMALLINT NOT NULL DEFAULT 0,
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_error TEXT,
    transaction_hash CHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (state_id) REFERENCES channel_states(id) ON DELETE CASCADE
);

CREATE INDEX idx_blockchain_actions_pending ON blockchain_actions(status, created_at) WHERE status = 0;
CREATE INDEX idx_blockchain_actions_state_id ON blockchain_actions(state_id);

-- Session keys table (LEGACY): Session keys with spending caps
CREATE TABLE session_keys (
    id SERIAL PRIMARY KEY,
    address VARCHAR NOT NULL UNIQUE,
    wallet_address VARCHAR NOT NULL,
    application VARCHAR NOT NULL,
    allowance JSONB,
    scope VARCHAR NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_session_keys_wallet_address ON session_keys(wallet_address);
CREATE UNIQUE INDEX idx_session_keys_unique_wallet_app ON session_keys(wallet_address, application);

-- +goose Down
DROP INDEX IF EXISTS idx_session_keys_unique_wallet_app;
DROP INDEX IF EXISTS idx_session_keys_wallet_address;
DROP TABLE IF EXISTS session_keys;
DROP INDEX IF EXISTS idx_blockchain_actions_state_id;
DROP INDEX IF EXISTS idx_blockchain_actions_pending;
DROP TABLE IF EXISTS blockchain_actions;
DROP INDEX IF EXISTS idx_contract_events_block;
DROP INDEX IF EXISTS contract_events_tx_log_chain_idx;
DROP TABLE IF EXISTS contract_events;
DROP INDEX IF EXISTS idx_app_ledger_v1_wallet;
DROP INDEX IF EXISTS idx_app_ledger_v1_account_asset;
DROP TABLE IF EXISTS app_ledger_v1;
DROP INDEX IF EXISTS idx_app_session_participants_v1_wallet;
DROP TABLE IF EXISTS app_session_participants_v1;
DROP INDEX IF EXISTS idx_app_sessions_v1_status;
DROP INDEX IF EXISTS idx_app_sessions_v1_application;
DROP TABLE IF EXISTS app_sessions_v1;
DROP INDEX IF EXISTS idx_transactions_to_comp;
DROP INDEX IF EXISTS idx_transactions_from_comp;
DROP INDEX IF EXISTS idx_transactions_from_to_type;
DROP INDEX IF EXISTS idx_transactions_to_account;
DROP INDEX IF EXISTS idx_transactions_from_account;
DROP INDEX IF EXISTS idx_transactions_type;
DROP TABLE IF EXISTS transactions;
DROP INDEX IF EXISTS idx_channel_states_escrow_channel_id;
DROP INDEX IF EXISTS idx_channel_states_home_channel_id;
DROP INDEX IF EXISTS idx_channel_states_epoch_version;
DROP INDEX IF EXISTS idx_channel_states_user_wallet_asset;
DROP INDEX IF EXISTS idx_channel_states_asset;
DROP INDEX IF EXISTS idx_channel_states_user_wallet;
DROP TABLE IF EXISTS channel_states;
DROP INDEX IF EXISTS idx_channels_status;
DROP INDEX IF EXISTS idx_channels_user_wallet;
DROP TABLE IF EXISTS channels;
