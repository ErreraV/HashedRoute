#!/usr/bin/env bash
# Deploy HashedRoute delivery chaincode to the Fabric test-network (always Org1 + Org2 + Org3 lifecycle).
#
# Prerequisites:
#   - fabric-samples test-network; channel mychannel; Org3 added (./addOrg3.sh up) so org3 crypto exists
#   - Docker + peers up
#
# Usage:
#   export FABRIC_TEST_NETWORK=.../fabric-samples/test-network
#   HASHEDRO_HOME=/path/to/HashedRoute $HASHEDRO_HOME/scripts/deploy-chaincode.sh
#
# Optional env:
#   CC_NAME   default: delivery
#   CC_VER    default: 1.0
#   SEQ       default: 1  (bump on upgrade)
#   CC_END_POLICY  override default OR('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')

set -euo pipefail

HASHEDRO_HOME="${HASHEDRO_HOME:-$(cd "$(dirname "$0")/.." && pwd)}"
CC_NAME="${CC_NAME:-delivery}"
CC_VER="${CC_VER:-1.0}"
SEQ="${SEQ:-1}"
CC_PATH_REL="${CC_PATH_REL:-$HASHEDRO_HOME/chaincode/delivery}"

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

# Default signature policy: all three test-network org peers may endorse.
DEFAULT_CC_ENDORSEMENT="OR('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')"
if [[ -n "${CC_END_POLICY:-}" ]]; then
  CCEPT="$CC_END_POLICY"
else
  CCEPT="$DEFAULT_CC_ENDORSEMENT"
fi

ORG3_CRYPTO="$TN/organizations/peerOrganizations/org3.example.com"
if [[ ! -d "$ORG3_CRYPTO" ]]; then
  echo "Org3 peer crypto is missing on disk: $ORG3_CRYPTO" >&2
  echo "This is normal right after make clean or before addOrg3 runs." >&2
  echo "Recreate orgs + channel + Org3, then deploy:" >&2
  echo "  make setup" >&2
  echo "or step by step:" >&2
  echo "  make fabric-network && make fabric-add-org3 && make fabric-deploy-chaincode" >&2
  exit 1
fi

cd "$TN"

echo "Deploying chaincode $CC_NAME $CC_VER (sequence $SEQ) from $CC_PATH_REL ..."
echo "Endorsement policy (override with CC_END_POLICY): $CCEPT"

export TEST_NETWORK_HOME="$TN"
export CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"
export CC_NAME
export CC_SRC_PATH="$CC_PATH_REL"
export CC_SRC_LANGUAGE="${CC_SRC_LANGUAGE:-go}"
export CC_VERSION="$CC_VER"
export CC_SEQUENCE="$SEQ"
export CC_INIT_FCN="${CC_INIT_FCN:-NA}"
export CC_END_POLICY="$CCEPT"
export CC_COLL_CONFIG="${CC_COLL_CONFIG:-NA}"
export DELAY="${DELAY:-${CLI_DELAY:-3}}"
export MAX_RETRY="${MAX_RETRY:-5}"
export VERBOSE="${VERBOSE:-false}"
bash "$HASHEDRO_HOME/scripts/fabric-lifecycle-three-org.sh"

echo "Done. Default channel: mychannel. Use CC_NAME=$CC_NAME in the API."
