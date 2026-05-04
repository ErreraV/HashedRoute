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

set -euo pipefail

HASHEDRO_HOME="${HASHEDRO_HOME:-$(cd "$(dirname "$0")/.." && pwd)}"
CC_NAME="${CC_NAME:-delivery}"
CC_VER="${CC_VER:-1.0}"
SEQ="${SEQ:-1}"
CC_PATH_REL="${CC_PATH_REL:-$HASHEDRO_HOME/chaincode/delivery}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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

cd "$TN"

echo "Deploying chaincode $CC_NAME $CC_VER (sequence $SEQ) from $CC_PATH_REL ..."

./network.sh deployCC \
  -ccn "$CC_NAME" \
  -ccp "$CC_PATH_REL" \
  -ccv "$CC_VER" \
  -ccl go \
  -ccs "$SEQ" \
  -ccep "OR('Org1MSP.peer','Org2MSP.peer')"

echo "Done. Default channel: mychannel. Use CC_NAME=$CC_NAME in the API."
