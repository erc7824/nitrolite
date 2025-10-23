-- +goose Up
-- +goose StatementBegin
DROP TABLE assets;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE assets (
    token VARCHAR NOT NULL,
    chain_id BIGINT NOT NULL,
    symbol VARCHAR NOT NULL,
    decimals BIGINT NOT NULL
);

ALTER TABLE assets ADD PRIMARY KEY(token, chain_id);
-- +goose StatementEnd
