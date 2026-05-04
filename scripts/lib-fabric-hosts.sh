#!/usr/bin/env bash
# HashedRoute is fixed to three Fabric org peers (Org1–Org3 on test-network).
# fabric-hosts.def may only tune API/web ports; columns and order are fixed (see fabric-hosts.def header).

# Default chaincode signature policy (all three peers may endorse).
HASHEDROUTE_CC_ENDORSEMENT_POLICY="OR('Org1MSP.peer','Org2MSP.peer','Org3MSP.peer')"

# Exit 0 if $1 is a valid fabric-hosts.def; else print to stderr and return 1.
fabric_hosts_validate_three() {
  local def="$1"
  [[ -f "$def" ]] || {
    echo "Not a file: $def" >&2
    return 1
  }
  local n=0 raw line key msp crypto_dir gateway_peer host_port _api _web
  while IFS= read -r raw || [[ -n "$raw" ]]; do
    line="${raw%%#*}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" ]] && continue
    n=$((n + 1))
    read -r key msp crypto_dir gateway_peer host_port _api _web <<<"$line"
    case $n in
    1)
      [[ "$key" == org1 && "$msp" == Org1MSP && "$crypto_dir" == org1.example.com && \
        "$gateway_peer" == peer0.org1.example.com && "$host_port" == 7051 ]] || {
        echo "fabric-hosts.def row 1 must be: org1 Org1MSP org1.example.com peer0.org1.example.com 7051 <api_port> <web_port>" >&2
        return 1
      }
      ;;
    2)
      [[ "$key" == org2 && "$msp" == Org2MSP && "$crypto_dir" == org2.example.com && \
        "$gateway_peer" == peer0.org2.example.com && "$host_port" == 9051 ]] || {
        echo "fabric-hosts.def row 2 must be: org2 Org2MSP org2.example.com peer0.org2.example.com 9051 <api_port> <web_port>" >&2
        return 1
      }
      ;;
    3)
      [[ "$key" == org3 && "$msp" == Org3MSP && "$crypto_dir" == org3.example.com && \
        "$gateway_peer" == peer0.org3.example.com && "$host_port" == 11051 ]] || {
        echo "fabric-hosts.def row 3 must be: org3 Org3MSP org3.example.com peer0.org3.example.com 11051 <api_port> <web_port>" >&2
        return 1
      }
      ;;
    *)
      echo "fabric-hosts.def must have exactly three data rows (no extra line): $line" >&2
      return 1
      ;;
    esac
  done <"$def"

  if [[ "$n" -ne 3 ]]; then
    echo "fabric-hosts.def must contain exactly three data rows (org1, org2, org3); found $n." >&2
    return 1
  fi
}
