#!/usr/bin/env bash
# Creates per-service RabbitMQ users for production.
# Run this once against the production broker using admin credentials.
# Usage: RABBIT_ADMIN_USER=admin RABBIT_ADMIN_PASS=<pass> RABBIT_HOST=localhost ./rabbitmq-setup-prod.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/.env"
if [[ -f "$ENV_FILE" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$ENV_FILE"
  set +a
fi

HOST="${RABBIT_HOST:-homenas.local}"
PORT="${RABBIT_MGMT_PORT:-15672}"
ADMIN_USER="${RABBIT_ADMIN_USER:-raino}"
ADMIN_PASS="${RABBIT_ADMIN_PASS:?RABBIT_ADMIN_PASS is required}"
VHOST="${RABBIT_VHOST:-tmf}"
VHOST_ENC="${VHOST/\//%2F}"  # URL-encode leading slash if vhost is "/"
BASE_URL="http://${HOST}:${PORT}/api"

create_vhost() {
  echo "Creating vhost: ${VHOST} (no-op if already exists)"
  # CloudAMQP: vhost is fixed to the account name — PUT returns 204 if created, 200 if exists.
  # We ignore failures here so the script works against both self-hosted and managed brokers.
  curl -s -o /dev/null -u "${ADMIN_USER}:${ADMIN_PASS}" \
    -H "Content-Type: application/json" \
    -X PUT "${BASE_URL}/vhosts/${VHOST_ENC}" || true
}

create_user() {
  local user="$1"
  local pass="$2"
  echo "Creating user: $user"
  curl -sf -u "${ADMIN_USER}:${ADMIN_PASS}" \
    -H "Content-Type: application/json" \
    -X PUT "${BASE_URL}/users/${user}" \
    -d "{\"password\":\"${pass}\",\"tags\":\"\"}"
  echo "Setting permissions for: $user on vhost ${VHOST}"
  # configure="" — exchanges and queues are pre-created by K8s CRDs, services must not create their own
  # write=".*"  — publish to any exchange
  # read=".*"   — consume from any queue
  curl -sf -u "${ADMIN_USER}:${ADMIN_PASS}" \
    -H "Content-Type: application/json" \
    -X PUT "${BASE_URL}/permissions/${VHOST_ENC}/${user}" \
    -d '{"configure":"","write":".*","read":".*"}'
}

create_vhost

create_user "tmf_party"         "${RABBITMQ_PARTY_PASS:?}"
create_user "tmf_customer"      "${RABBITMQ_CUSTOMER_PASS:?}"
create_user "tmf_catalog"       "${RABBITMQ_CATALOG_PASS:?}"
create_user "tmf_qualification" "${RABBITMQ_QUALIFICATION_PASS:?}"
create_user "tmf_cart"          "${RABBITMQ_CART_PASS:?}"
create_user "tmf_pocv"          "${RABBITMQ_POCV_PASS:?}"
create_user "tmf_bff"           "${RABBITMQ_BFF_PASS:?}"

echo "Done."
