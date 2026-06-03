#!/usr/bin/env bash
set -euo pipefail

if [ "$#" -ne 1 ]; then
  echo "usage: scripts/restore-local-db.sh backups/example.dump" >&2
  exit 1
fi

backup_file="$1"
if [ ! -f "$backup_file" ]; then
  echo "backup file not found: $backup_file" >&2
  exit 1
fi

sg docker -c "docker compose exec -T postgres pg_restore -U form_invoice_generator -d form_invoice_generator --clean --if-exists" < "$backup_file"
