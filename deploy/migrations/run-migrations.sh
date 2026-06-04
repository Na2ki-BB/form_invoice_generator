#!/usr/bin/env sh
set -eu

if [ -z "${DATABASE_URL:-}" ]; then
  echo "DATABASE_URL is required." >&2
  exit 1
fi

if [ ! -d /migrations ]; then
  echo "Migration directory was not found." >&2
  exit 1
fi

echo "Starting database migrations."

for file in /migrations/*.sql; do
  if [ ! -f "$file" ]; then
    echo "No migration files were found." >&2
    exit 1
  fi

  echo "Applying $(basename "$file")"
  psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f "$file"
done

echo "Database migrations completed."
