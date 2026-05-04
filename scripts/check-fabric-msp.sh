#!/usr/bin/env bash
# Ensure User1 MSP material exists for org1/org2/org3 under fabric-samples test-network (before make up).
set -euo pipefail

: "${FABRIC_SAMPLES_ROOT:?FABRIC_SAMPLES_ROOT must be set (e.g. via Makefile / .env)}"

TN="${FABRIC_SAMPLES_ROOT}/test-network"
err=0
for o in org1.example.com org2.example.com org3.example.com; do
  msp="$TN/organizations/peerOrganizations/$o/users/User1@$o/msp"
  if [[ ! -d "$msp" ]]; then
    echo "Missing MSP: $msp" >&2
    echo "  Run make setup (or make fabric-network && make fabric-add-org3) first." >&2
    err=1
  fi
done
exit "$err"
