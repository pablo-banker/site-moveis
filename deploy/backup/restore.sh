#!/usr/bin/env bash
set -euo pipefail
: "${RESTORE_DATABASE_URL:?Set a separate empty restore database}"
: "${RESTORE_AGE_IDENTITY:?Set path to offline age private identity}"
: "${CONFIRM_RESTORE:?Set CONFIRM_RESTORE=isolated-target after checking the destination}"
[[ "$CONFIRM_RESTORE" == "isolated-target" ]] || exit 1
restore_archive="${1:?Pass the encrypted archive path}"
# No DROP/CLEAN: existing conflicting objects stop the transaction.
age --decrypt --identity "$RESTORE_AGE_IDENTITY" "$restore_archive" | pg_restore --dbname="$RESTORE_DATABASE_URL" --single-transaction --exit-on-error --no-owner --no-privileges
printf 'Restore completed. Validate counts and application flows in the isolated database.\n'
