-- +goose Up
-- +goose StatementBegin
CREATE TABLE signers (
    signer VARCHAR PRIMARY KEY,
    wallet VARCHAR NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE signers;
-- +goose StatementEnd
