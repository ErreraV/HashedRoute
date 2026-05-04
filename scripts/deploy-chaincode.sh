#!/usr/bin/env bash
# Deploy HashedRoute delivery chaincode to the Fabric test-network.
#
# Prerequisites:
#   - Cloned https://github.com/hyperledger/fabric-samples (release-2.5 branch recommended)
#   - Docker + test network up: ./network.sh up createChannel
#
# Usage (from fabric-samples/test-network):
#   export HASHEDRO_HOME=/path/to/HashedRoute
#   $HASHEDRO_HOME/scripts/deploy-chaincode.sh
#
# Optional env:
#   CC_NAME   default: delivery
#   CC_VER    default: 1.0
#   SEQ       default: 1  (bump on upgrade)
#   FABRIC_HOSTS_DEF  default: $HASHEDRO_HOME/fabric-hosts.def — MSP column sets default endorsement OR(...)
#   CC_END_POLICY  optional override (skip parsing fabric-hosts.def)

set -euo pipefail

HASHEDRO_HOME="${HASHEDRO_HOME:-$(cd "$(dirname "$0")/.." && pwd)}"
CC_NAME="${CC_NAME:-delivery}"
CC_VER="${CC_VER:-1.0}"
SEQ="${SEQ:-1}"
CC_PATH_REL="${CC_PATH_REL:-$HASHEDRO_HOME/chaincode/delivery}"
DEF="${FABRIC_HOSTS_DEF:-$HASHEDRO_HOME/fabric-hosts.def}"

TN="$(cd "${FABRIC_TEST_NETWORK:-}" 2>/dev/null && pwd || true)"

if [[ -z "$TN" || ! -f "$TN/network.sh" ]]; then
  echo "Set FABRIC_TEST_NETWORK to the path of fabric-samples/test-network (contains network.sh)." >&2
  echo "Example: export FABRIC_TEST_NETWORK=~/fabric-samples/test-network" >&2
  exit 1
fi

if [[ ! -d "$CC_PATH_REL" ]]; then
  echo "Chaincode path not found: $CC_PATH_REL" >&2
  exit 1
fi

if [[ ! -f "$DEF" ]]; then
  echo "Missing $DEF — set FABRIC_HOSTS_DEF or create fabric-hosts.def" >&2
  exit 1
fi

# shellcheck source=/dev/null
. "$(dirname "$0")/lib-fabric-hosts.sh"

if [[ -n "${CC_END_POLICY:-}" ]]; then
  CCEPT="$CC_END_POLICY"
else
  CCEPT="$(fabric_hosts_endorsement_or "$DEF")" || exit 1
fi

cd "$TN"

echo "Deploying chaincode $CC_NAME $CC_VER (sequence $SEQ) from $CC_PATH_REL ..."
echo "Endorsement policy (from ${DEF} unless CC_END_POLICY set): $CCEPT"

./network.sh deployCC \
  -ccn "$CC_NAME" \
  -ccp "$CC_PATH_REL" \
  -ccv "$CC_VER" \
  -ccl go \
  -ccs "$SEQ" \
  -ccep "$CCEPT"

echo "Done. Default channel: mychannel. Use CC_NAME=$CC_NAME in the API."
