#!/usr/bin/env bash
# Create missing peerOrganizations material for rows in fabric-hosts.def when the stock
# fabric-samples workflow can do it (today: org3.example.com via addOrg3.sh).
#
# Env:
#   FABRIC_HOSTS_DEF      default: HASHEDRO_HOME/fabric-hosts.def
#   FABRIC_TEST_NETWORK   required: .../fabric-samples/test-network
#   HASHEDRO_HOME         repo root

set -euo pipefail

ROOT="${HASHEDRO_HOME:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
DEF="${FABRIC_HOSTS_DEF:-$ROOT/fabric-hosts.def}"
TN="$(cd "${FABRIC_TEST_NETWORK:-}" 2>/dev/null && pwd || true)"

if [[ -z "$TN" || ! -f "$TN/network.sh" ]]; then
  echo "Set FABRIC_TEST_NETWORK to fabric-samples/test-network (contains network.sh)." >&2
  exit 1
fi

if [[ ! -f "$DEF" ]]; then
  echo "Missing $DEF" >&2
  exit 1
fi

user_msp_dir() {
  local crypto_dir="$1"
  echo "$TN/organizations/peerOrganizations/${crypto_dir}/users/User1@${crypto_dir}/msp"
}

missing=()

while IFS= read -r raw || [[ -n "$raw" ]]; do
  line="${raw%%#*}"
  line="${line#"${line%%[![:space:]]*}"}"
  line="${line%"${line##*[![:space:]]}"}"
  [[ -z "$line" ]] && continue

  read -r key _msp crypto_dir _g _h _a _w <<<"$line"
  if [[ -z "${crypto_dir:-}" ]]; then
    echo "Invalid line in $DEF (need CRYPTO_PEER_DIR): $raw" >&2
    exit 1
  fi

  path="$(user_msp_dir "$crypto_dir")"
  [[ -d "$path" ]] && continue

  dup=0
  if ((${#missing[@]})); then
    for d in "${missing[@]}"; do
      if [[ "$d" == "$crypto_dir" ]]; then
        dup=1
        break
      fi
    done
  fi
  [[ "$dup" -eq 0 ]] && missing+=("$crypto_dir")
done <"$DEF"

if [[ ${#missing[@]} -eq 0 ]]; then
  exit 0
fi

need_base=0
need_org3=0
unsupported=()
for d in "${missing[@]}"; do
  case "$d" in
    org1.example.com|org2.example.com)
      need_base=1
      ;;
    org3.example.com)
      need_org3=1
      ;;
    *)
      unsupported+=("$d")
      ;;
  esac
done

if ((${#unsupported[@]})); then
  echo "No automatic provisioning for peer org(s): ${unsupported[*]}" >&2
  echo "  HashedRoute only auto-runs fabric-samples test-network/addOrg3 for org3.example.com." >&2
  echo "  Add crypto under $TN/organizations/peerOrganizations/ yourself, or remove those rows from $DEF." >&2
  exit 1
fi

if [[ "$need_base" -ne 0 ]]; then
  echo "Missing MSP for: ${missing[*]}" >&2
  echo "  Org1/Org2 come from the test network — run: make fabric-network   or   make fabric-setup" >&2
  echo "  Then re-run: make up" >&2
  exit 1
fi

if [[ "$need_org3" -ne 0 ]]; then
  add_script="$TN/addOrg3/addOrg3.sh"
  if [[ ! -f "$add_script" ]]; then
    echo "Missing $add_script (full fabric-samples clone required to add Org3)." >&2
    exit 1
  fi
  org1_msp="$(user_msp_dir org1.example.com)"
  if [[ ! -d "$org1_msp" ]]; then
    echo "Org3 add requested but Org1 MSP is missing — start the network first:" >&2
    echo "  make fabric-network   or   make fabric-setup" >&2
    exit 1
  fi
  echo "Provisioning org3.example.com via fabric-samples addOrg3 (channel mychannel must exist) ..."
  (cd "$TN/addOrg3" && bash ./addOrg3.sh up)
  echo ""
  echo "If chaincode was deployed before Org3 joined, redeploy with a higher SEQ / matching policy, e.g.:"
  echo "  make fabric-deploy-chaincode SEQ=2"
fi
