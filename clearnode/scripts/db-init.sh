#!/usr/bin/env bash
set -e

echo 'Waiting for database to be ready...'

until pg_isready -h database -p 5432 -U ${POSTGRES_USER}; do
  echo 'Waiting for database connection...'
  sleep 2
done

psql -h database -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "
  CREATE TABLE IF NOT EXISTS goose_db_version (
    id serial PRIMARY KEY,
    version_id int8 NOT NULL,
    is_applied boolean NOT NULL DEFAULT true,
    tstamp timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
  );
"

echo 'Checking for migration files...'

echo 'Running pending migrations...'
for migration in /migrations/*.sql; do
  filename=$(basename $migration) &&
  version=$(echo $filename | grep -o '^[0-9]\+') &&

  if ! psql -h database -U ${POSTGRES_USER} -d ${POSTGRES_DB} -tAc "SELECT 1 FROM goose_db_version WHERE version_id = $version" | grep -q 1; then
    echo "Applying migration: $filename"
    sed -n '/^-- +goose Up/,/^-- +goose Down/p' $migration |
    grep -v '^-- +goose' |
    psql -h database -U ${POSTGRES_USER} -d ${POSTGRES_DB} &&
    psql -h database -U ${POSTGRES_USER} -d ${POSTGRES_DB} -c "INSERT INTO goose_db_version (version_id) VALUES ($version);" &&
    echo "Successfully applied: $filename"
  fi
done

echo '--- Starting Token Seeding Process ---'
export PSQL="psql -h database -U ${POSTGRES_USER} -d ${POSTGRES_DB}"
chmod +x /scripts/seed-tokens.sh
/scripts/seed-tokens.sh || echo "Token seeding failed, but continuing..."
echo '--- Token Seeding Process Complete ---'
