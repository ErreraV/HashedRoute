#!/usr/bin/env bash
# Chaincode lifecycle for test-network with Org1+2+3 (extends fabric-samples/scripts/deployCC.sh).
# Run from repo via deploy-chaincode.sh, or with cwd / TEST_NETWORK_HOME = fabric-samples/test-network.
#
# Required env: CC_NAME CC_SRC_PATH CC_SRC_LANGUAGE CC_VERSION CC_SEQUENCE
# Optional: CHANNEL_NAME (default mychannel), CC_INIT_FCN (default NA), CC_END_POLICY, CC_COLL_CONFIG,
#           DELAY MAX_RETRY VERBOSE TEST_NETWORK_HOME

# Do not use `set -e` or `set -o pipefail` here: fabric-samples ccutils.sh uses `let rc=0`, and in bash `let`
# returns exit status 1 when the expression is 0, which would abort the script under errexit. deployCC.sh
# does not enable -e. Do not use `set -u`: ccutils may reference unset vars in edge paths.

TN="${TEST_NETWORK_HOME:-}"
if [[ -z "$TN" || ! -f "$TN/network.sh" ]]; then
  echo "TEST_NETWORK_HOME must be fabric-samples/test-network (directory with network.sh)." >&2
  exit 1
fi

cd "$TN"
export TEST_NETWORK_HOME="$TN"

# Same as network.sh: peer/configtxgen live under fabric-samples/bin (sibling of test-network/).
FABRIC_SAMPLES_BIN="$(cd "$TN/.." && pwd)/bin"
export PATH="${FABRIC_SAMPLES_BIN}:${PATH}"
if ! command -v peer >/dev/null 2>&1; then
  echo "peer not in PATH. Install Fabric binaries (e.g. make install) or ensure exists: ${FABRIC_SAMPLES_BIN}/peer" >&2
  exit 1
fi

CHANNEL_NAME="${CHANNEL_NAME:-mychannel}"
CC_INIT_FCN="${CC_INIT_FCN:-NA}"
CC_END_POLICY="${CC_END_POLICY:-NA}"
CC_COLL_CONFIG="${CC_COLL_CONFIG:-NA}"
DELAY="${DELAY:-3}"
MAX_RETRY="${MAX_RETRY:-5}"
VERBOSE="${VERBOSE:-false}"

INIT_REQUIRED="--init-required"
if [ "$CC_INIT_FCN" = "NA" ]; then
  INIT_REQUIRED=""
fi

if [ "$CC_END_POLICY" = "NA" ]; then
  CC_END_POLICY=""
else
  CC_END_POLICY="--signature-policy $CC_END_POLICY"
fi

if [ "$CC_COLL_CONFIG" = "NA" ]; then
  CC_COLL_CONFIG=""
else
  CC_COLL_CONFIG="--collections-config $CC_COLL_CONFIG"
fi

# peer runs in child processes (packageCC.sh, etc.); must be exported. network.sh does export before deployCC.
export FABRIC_CFG_PATH="$(cd "$TN/.." && pwd)/config"

# shellcheck source=/dev/null
. scripts/utils.sh

function checkPrereqs() {
  jq --version >/dev/null 2>&1 || {
    errorln "jq command not found..."
    errorln "https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html"
    exit 1
  }
}

checkPrereqs

# shellcheck source=/dev/null
. scripts/envVar.sh
# shellcheck source=/dev/null
. scripts/ccutils.sh

# Override fabric-samples installChaincode: if package already on peer, upstream never sets $res and verifyResult breaks.
installChaincode() {
  ORG=$1
  setGlobals $ORG
  set -x
  peer lifecycle chaincode queryinstalled --output json | jq -r 'try (.installed_chaincodes[].package_id)' | grep ^${PACKAGE_ID}$ >&log.txt
  if test $? -ne 0; then
    peer lifecycle chaincode install ${CC_NAME}.tar.gz >&log.txt
    res=$?
  else
    res=0
  fi
  { set +x; } 2>/dev/null
  cat log.txt
  verifyResult $res "Chaincode installation on peer0.org${ORG} has failed"
  successln "Chaincode is installed on peer0.org${ORG}"
}

println "executing 3-org lifecycle with the following"
println "- CHANNEL_NAME: ${C_GREEN}${CHANNEL_NAME}${C_RESET}"
println "- CC_NAME: ${C_GREEN}${CC_NAME}${C_RESET}"
println "- CC_SRC_PATH: ${C_GREEN}${CC_SRC_PATH}${C_RESET}"
println "- CC_SRC_LANGUAGE: ${C_GREEN}${CC_SRC_LANGUAGE}${C_RESET}"
println "- CC_VERSION: ${C_GREEN}${CC_VERSION}${C_RESET}"
println "- CC_SEQUENCE: ${C_GREEN}${CC_SEQUENCE}${C_RESET}"
println "- DELAY: ${C_GREEN}${DELAY}${C_RESET}"
println "- MAX_RETRY: ${C_GREEN}${MAX_RETRY}${C_RESET}"
println "- VERBOSE: ${C_GREEN}${VERBOSE}${C_RESET}"

./scripts/packageCC.sh "$CC_NAME" "$CC_SRC_PATH" "$CC_SRC_LANGUAGE" "$CC_VERSION"

PACKAGE_ID=$(peer lifecycle chaincode calculatepackageid "${CC_NAME}.tar.gz")

infoln "Installing chaincode on peer0.org1..."
installChaincode 1
infoln "Installing chaincode on peer0.org2..."
installChaincode 2
infoln "Installing chaincode on peer0.org3..."
installChaincode 3

resolveSequence

queryInstalled 1

approveForMyOrg 1
checkCommitReadiness 1 "\"Org1MSP\": true" "\"Org2MSP\": false" "\"Org3MSP\": false"
checkCommitReadiness 2 "\"Org1MSP\": true" "\"Org2MSP\": false" "\"Org3MSP\": false"
checkCommitReadiness 3 "\"Org1MSP\": true" "\"Org2MSP\": false" "\"Org3MSP\": false"

approveForMyOrg 2
checkCommitReadiness 1 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": false"
checkCommitReadiness 2 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": false"
checkCommitReadiness 3 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": false"

approveForMyOrg 3
checkCommitReadiness 1 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": true"
checkCommitReadiness 2 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": true"
checkCommitReadiness 3 "\"Org1MSP\": true" "\"Org2MSP\": true" "\"Org3MSP\": true"

commitChaincodeDefinition 1 2 3

queryCommitted 1
queryCommitted 2
queryCommitted 3

if [ "$CC_INIT_FCN" = "NA" ]; then
  infoln "Chaincode initialization is not required"
else
  chaincodeInvokeInit 1 2 3
fi

exit 0
