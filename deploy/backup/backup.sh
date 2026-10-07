#!/usr/bin/env bash
set -euo pipefail
umask 077
: "${BACKUP_DATABASE_URL:?Set dedicated backup database connection}"
: "${BACKUP_AGE_RECIPIENT:?Set age public recipient}"
: "${BACKUP_DIRECTORY:=/backups}"
mkdir -p "$BACKUP_DIRECTORY"
backup_path="$BACKUP_DIRECTORY/forma-$(date -u +%Y%m%dT%H%M%SZ).dump.age"
backup_temp="$backup_path.partial"
trap 'rm -f "$backup_temp"' EXIT
pg_dump --dbname="$BACKUP_DATABASE_URL" --format=custom --no-owner --no-privileges | age --recipient "$BACKUP_AGE_RECIPIENT" > "$backup_temp"
mv "$backup_temp" "$backup_path"
printf 'Encrypted backup written: %s\n' "$backup_path"
