-- +goose Up
-- +goose StatementBegin

ALTER TABLE channels ADD COLUMN state TEXT;
ALTER TABLE channels ADD COLUMN server_state_signature TEXT;
ALTER TABLE channels ADD COLUMN user_state_signature TEXT;

UPDATE channels
SET state = json_build_object(
    'intent', 1,
    'version', version,
    'state_data', '',
    'allocations', '[]'::jsonb
)::text;

ALTER TABLE channels ALTER COLUMN state SET NOT NULL;

ALTER TABLE channels DROP COLUMN version;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE channels ADD COLUMN version BIGINT NOT NULL DEFAULT 1;
UPDATE channels
SET version = COALESCE((state::json ->> 'version')::bigint, 1);

ALTER TABLE channels DROP COLUMN state;
ALTER TABLE channels DROP COLUMN server_state_signature;
ALTER TABLE channels DROP COLUMN user_state_signature;
-- +goose StatementEnd
