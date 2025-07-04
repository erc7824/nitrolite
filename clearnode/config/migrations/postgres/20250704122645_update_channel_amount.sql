-- +goose Up
-- This section migrates the 'channels' table by renaming the 'amount'
-- column to 'raw_amount' and changing its data type from BIGINT to a
-- high-precision DECIMAL(64,18).

-- +goose StatementBegin
ALTER TABLE channels
RENAME COLUMN amount TO raw_amount;

ALTER TABLE channels
ALTER COLUMN raw_amount TYPE DECIMAL(64,18)
USING raw_amount::DECIMAL(64,18);
-- +goose StatementEnd

-- +goose Down
-- This section reverts the migration. It changes the 'raw_amount' column
-- back to 'amount' and converts its data type from DECIMAL(64,18) to BIGINT.
--
-- WARNING: This is a potentially lossy conversion. Any fractional data
-- in the decimal 'raw_amount' will be truncated (e.g., 123.45 becomes 123).
--
-- +goose StatementBegin
ALTER TABLE channels
ALTER COLUMN raw_amount TYPE BIGINT
USING raw_amount::BIGINT;

ALTER TABLE channels
RENAME COLUMN raw_amount TO amount;
-- +goose StatementEnd
