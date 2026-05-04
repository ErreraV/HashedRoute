#!/usr/bin/env bash
# Shared helpers for fabric-hosts.def (same 7-column format as gen/check scripts).
# Bash 3.2–compatible (no namerefs).

# Print unique MSP IDs from DEF (first occurrence order, one per line). Returns nonzero on bad lines.
fabric_hosts_unique_msps() {
  local def="$1"
  [[ -f "$def" ]] || {
    echo "Not a file: $def" >&2
    return 1
  }
  local raw line key msp _c _g _h _a _w m seen
  local -a acc=()

  while IFS= read -r raw || [[ -n "$raw" ]]; do
    line="${raw%%#*}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -z "$line" ]] && continue
    read -r key msp _c _g _h _a _w <<<"$line"
    if [[ -z "${msp:-}" ]]; then
      echo "Invalid line in $def (need MSP_ID): $raw" >&2
      return 1
    fi
    seen=0
    if ((${#acc[@]})); then
      for m in "${acc[@]}"; do
        if [[ "$m" == "$msp" ]]; then
          seen=1
          break
        fi
      done
    fi
    [[ "$seen" -eq 0 ]] && acc+=("$msp")
  done <"$def"

  if ((${#acc[@]})); then
    for m in "${acc[@]}"; do
      printf '%s\n' "$m"
    done
  fi
}

# Print OR('MSPid.peer',...) from DEF's MSP column (unique, order preserved).
fabric_hosts_endorsement_or() {
  local def="$1"
  local msps=() m parts=() tmp

  tmp="$(mktemp "${TMPDIR:-/tmp}/hashedroute-hosts.XXXXXX")"
  if ! fabric_hosts_unique_msps "$def" >"$tmp"; then
    rm -f "$tmp"
    return 1
  fi
  while IFS= read -r m || [[ -n "$m" ]]; do
    [[ -n "${m:-}" ]] && msps+=("$m")
  done <"$tmp"
  rm -f "$tmp"

  if [[ ${#msps[@]} -eq 0 ]]; then
    echo "No MSP IDs found in $def" >&2
    return 1
  fi

  for m in "${msps[@]}"; do
    parts+=("'${m}.peer'")
  done
  local IFS=','
  printf 'OR(%s)' "${parts[*]}"
}
