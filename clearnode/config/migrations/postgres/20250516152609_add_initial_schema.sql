-- +goose Up
-- TODO: need to discuss the primary key
CREATE TABLE channels (
    channel_id VARCHAR PRIMARY KEY,
    chain_id BIGINT NOT NULL,
    token VARCHAR NOT NULL,
    wallet VARCHAR NOT NULL,
    participant VARCHAR NOT NULL,
    amount BIGINT NOT NULL,
    status VARCHAR NOT NULL,
    challenge BIGINT DEFAULT 0,
    nonce BIGINT DEFAULT 0,
    version BIGINT DEFAULT 0,
    adjudicator VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE assets (
    token VARCHAR NOT NULL,
    chain_id BIGINT NOT NULL,
    symbol VARCHAR NOT NULL,
    decimals BIGINT NOT NULL
);

ALTER TABLE assets ADD PRIMARY KEY(token, chain_id);

CREATE TABLE ledger (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR NOT NULL,
    account_type BIGINT NOT NULL,
    asset_symbol VARCHAR NOT NULL,
    wallet VARCHAR NOT NULL,
    credit DECIMAL(64,18) NOT NULL,
    debit DECIMAL(64,18)NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE app_sessions (
    id SERIAL PRIMARY KEY,
    protocol VARCHAR NOT NULL DEFAULT 'NitroRPC/0.2',
    session_id VARCHAR NOT NULL UNIQUE,
    challenge BIGINT,
    nonce BIGINT NOT NULL,
    participants TEXT[] NOT NULL,
    weights INTEGER[],
    quorum BIGINT DEFAULT 100,
    version BIGINT DEFAULT 1,
    status VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE rpc_store (
    id SERIAL PRIMARY KEY,
    sender VARCHAR(255) NOT NULL,
    req_id BIGINT NOT NULL,
    method VARCHAR(255) NOT NULL,
    params TEXT NOT NULL,
    timestamp BIGINT NOT NULL,
    req_sig TEXT[],
    response TEXT NOT NULL,
    res_sig TEXT[]
);



-- +goose Down
-- +goose StatementBegin
DROP TABLE channels;
DROP TABLE assets;
DROP TABLE ledger;
DROP TABLE app_sessions;
DROP TABLE rpc_store;
-- +goose StatementEnd
