-- +goose Up
-- This section migrates the 'amount' column in the 'channels' table
-- from its original BIGINT type to a high-precision DECIMAL(64,18) type.

-- +goose StatementBegin
ALTER TABLE channels
ALTER COLUMN amount TYPE DECIMAL(64,18)
USING amount::DECIMAL(64,18);
-- +goose StatementEnd

-- +goose Down
-- This section reverts the migration, changing the 'amount' column
-- back from DECIMAL(64,18) to BIGINT.
--
-- WARNING: This is a potentially lossy conversion. Any fractional data
-- in the decimal 'amount' will be truncated (e.g., 123.45 becomes 123).
--
-- +goose StatementBegin
ALTER TABLE channels
ALTER COLUMN amount TYPE BIGINT
USING amount::BIGINT;
-- +goose StatementEnd
