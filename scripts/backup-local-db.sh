#!/usr/bin/env bash
set -euo pipefail

backup_dir="${1:-backups}"
mkdir -p "$backup_dir"
backup_file="$backup_dir/form_invoice_generator_$(date +%Y%m%d_%H%M%S).dump"

sg docker -c "docker compose exec -T postgres pg_dump -U form_invoice_generator -d form_invoice_generator -Fc" > "$backup_file"
echo "$backup_file"
