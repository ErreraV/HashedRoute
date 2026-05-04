#!/usr/bin/env bash
# Submit mocked DeliveryContract transactions via Org1 Admin (test-network peer CLI).
#
# Prerequisites: test network up, channel created, chaincode deployed (e.g. make fabric-setup).
#
# Environment (optional):
#   FABRIC_TEST_NETWORK  path to fabric-samples/test-network (default: from HASHEDRO_HOME Makefile flow)
#   CHANNEL_NAME         default: mychannel
#   CC_NAME              default: delivery
#
# Usage:
#   HASHEDRO_HOME=/path/to/HashedRoute FABRIC_TEST_NETWORK=.../fabric-samples/test-network ./seed-mock-ledger.sh

set -euo pipefail

HASHEDRO_HOME="${HASHEDRO_HOME:-$(cd "$(dirname "$0")/.." && pwd)}"
CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"
CC_NAME="${CC_NAME:-delivery}"

TN="$(cd "${FABRIC_TEST_NETWORK:-}" 2>/dev/null && pwd || true)"
if [[ -z "$TN" || ! -f "$TN/network.sh" ]]; then
  echo "Set FABRIC_TEST_NETWORK to fabric-samples/test-network (contains network.sh)." >&2
  exit 1
fi

SAMPLES_ROOT="$(cd "$(dirname "$TN")" && pwd)"
export PATH="${SAMPLES_ROOT}/bin:${PATH}"
export FABRIC_CFG_PATH="${SAMPLES_ROOT}/config"

cd "$TN"
# envVar.sh references these under set -u; give safe defaults.
export OVERRIDE_ORG="${OVERRIDE_ORG:-}"
export VERBOSE="${VERBOSE:-false}"
# envVar.sh expects test-network as PWD; TEST_NETWORK_HOME defaults to PWD.
# shellcheck source=/dev/null
. "${TN}/scripts/envVar.sh"

setGlobals 1

if ! command -v peer >/dev/null 2>&1; then
  echo "peer not found. Install Fabric binaries (e.g. make fabric-install-hyperledger) and ensure \$PATH includes fabric-samples/bin." >&2
  exit 1
fi

invoke() {
  # $1 = chaincode call JSON (-c payload)
  local payload="$1"
  peer chaincode invoke \
    -o localhost:7050 \
    --ordererTLSHostnameOverride orderer.example.com \
    --tls \
    --cafile "${ORDERER_CA}" \
    -C "${CHANNEL_NAME}" \
    -n "${CC_NAME}" \
    --peerAddresses localhost:7051 \
    --tlsRootCertFiles "${PEER0_ORG1_CA}" \
    --waitForEvent \
    -c "${payload}"
}

echo "Seeding ${CC_NAME} on ${CHANNEL_NAME} with mock shipments (Org1) ..."

invoke '{"function":"CreateShipment","Args":["MOCK-SHP-001","New York NY","London UK","Acme Trading Co.","BlueHarbor Lines"]}'
invoke '{"function":"CreateShipment","Args":["MOCK-SHP-002","Sao Paulo BR","Lisbon PT","Globex Retail","Iberia Cargo"]}'
invoke '{"function":"CreateShipment","Args":["MOCK-SHP-003","Chicago IL","Seattle WA","Northwind Foods","Pacific Overland"]}'

invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-001","PICKED_UP","Picked up at origin dock"]}'
invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-001","IN_TRANSIT","Departed origin hub"]}'
invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-001","DELIVERED","Signed by consignee"]}'

invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-002","PICKED_UP","Staging scan complete"]}'

invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-003","PICKED_UP","Loaded at Chicago"]}'
invoke '{"function":"UpdateStatus","Args":["MOCK-SHP-003","IN_TRANSIT","En route to Seattle"]}'

echo "Done. MOCK-SHP-001=DELIVERED, MOCK-SHP-002=PICKED_UP, MOCK-SHP-003=IN_TRANSIT."
