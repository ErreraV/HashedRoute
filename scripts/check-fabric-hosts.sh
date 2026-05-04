#!/usr/bin/env bash
# Verify MSP material exists for each stack in fabric-hosts.def.
set -euo pipefail

ROOT="${HASHEDRO_HOME:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
DEF="${FABRIC_HOSTS_DEF:-$ROOT/fabric-hosts.def}"

: "${FABRIC_SAMPLES_ROOT:?FABRIC_SAMPLES_ROOT must be set (e.g. via Makefile / .env)}"

if [[ ! -f "$DEF" ]]; then
  echo "Missing $DEF" >&2
  exit 1
fi

err=0
while IFS= read -r raw || [[ -n "$raw" ]]; do
  line="${raw%%#*}"
  line="${line#"${line%%[![:space:]]*}"}"
  line="${line%"${line##*[![:space:]]}"}"
  [[ -z "$line" ]] && continue
  read -r key msp crypto_dir gateway_peer host_port api_port web_port <<<"$line"
  user_dir="User1@${crypto_dir}"
  msp_root="${FABRIC_SAMPLES_ROOT}/test-network/organizations/peerOrganizations/${crypto_dir}/users/${user_dir}/msp"
  if [[ ! -d "$msp_root" ]]; then
    echo "Missing MSP for stack '$key': $msp_root" >&2
    echo "  (fabric-hosts.def line was: $line)" >&2
    err=1
  fi
done <"$DEF"

exit "$err"
