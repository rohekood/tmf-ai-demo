#!/usr/bin/env bash
# Creates per-service Valkey ACL users for production.
# Each user can only access keys with their service prefix.
# Run once against the production Valkey instance.
# Usage: VALKEY_ADMIN_PASS=<admin-pass> ./valkey-setup-prod.sh
#
# Per-service key prefixes:
#   qualification  ->  qualification:*

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  set +a
fi

HOST="${VALKEY_HOST:-192.168.10.254}"
PORT="${VALKEY_PORT:-6379}"
ADMIN_PASS="${VALKEY_ADMIN_PASS:?VALKEY_ADMIN_PASS is required}"

vcli() {
  valkey-cli -h "$HOST" -p "$PORT" -a "$ADMIN_PASS" --no-auth-warning "$@"
}

create_user() {
  local user="$1"
  local pass="$2"
  local key_pattern="$3"
  echo "Creating Valkey ACL user: $user (keys: ${key_pattern})"
  # on  — account enabled
  # ><pass>  — set password
  # ~<pattern>  — allow keys matching pattern only
  # &*  — allow all pub/sub channels
  # +@read +@write +@connection  — allow read, write, and connection commands
  vcli ACL SETUSER "$user" on ">$pass" "~${key_pattern}" "&*" "+@read" "+@write" "+@connection"
}

create_user "valkey_qualification" "${VALKEY_QUALIFICATION_PASS:?}" "qualification:*"

echo "Saving ACL to disk..."
vcli ACL SAVE

echo "Done. Active users:"
vcli ACL LIST
