#!/usr/bin/env bash
set -euo pipefail

TABLE="${TABLE:-assets}"
PSQL="${PSQL:-psql}"
REQUIRED_PROPS=("TOKEN" "CHAIN_ID" "SYMBOL" "DECIMALS")
UPSERT="${UPSERT:-1}"
sqlq() {
  local s="${1//\'/\'\'}"
  printf "'%s'" "$s"
}
is_uint() { [[ "${1:-}" =~ ^[0-9]+$ ]]; }

declare -A V
declare -A IDX
while IFS= read -r name; do
  if [[ "$name" =~ ^TOKEN_([A-Za-z0-9]+)_(TOKEN|ADDRESS|CHAIN_ID|SYMBOL|DECIMALS)$ ]]; then
    idx="${BASH_REMATCH[1]}"
    prop="${BASH_REMATCH[2]}"
    [[ "$prop" == "ADDRESS" ]] && prop="TOKEN"
    val="${!name-}"
    V["$idx|$prop"]="$val"
    IDX["$idx"]=1
  fi
done < <(compgen -e)

if [ "${#IDX[@]}" -eq 0 ]; then
  echo "No TOKEN_<index>_{TOKEN|ADDRESS|CHAIN_ID|SYMBOL|DECIMALS} variables found." >&2
  exit 1
fi

sql="BEGIN;\n"
for idx in "${!IDX[@]}"; do
  missing=()
  for p in "${REQUIRED_PROPS[@]}"; do
    [[ -n "${V["$idx|$p"]+x}" ]] || missing+=("$p")
  done
  if ((${#missing[@]})); then
    echo "Skip index '$idx': missing ${missing[*]}" >&2
    continue
  fi

  token="${V["$idx|TOKEN"]}"
  chain_id="${V["$idx|CHAIN_ID"]}"
  symbol="${V["$idx|SYMBOL"]}"
  decimals="${V["$idx|DECIMALS"]}"

  is_uint "$chain_id" || { echo "Skip $idx: CHAIN_ID not uint ($chain_id)"; continue; }
  is_uint "$decimals" || { echo "Skip $idx: DECIMALS not uint ($decimals)"; continue; }

  if [ "$UPSERT" = "1" ]; then
    stmt=$(
      cat <<SQL
INSERT INTO ${TABLE} (token, chain_id, symbol, decimals)
VALUES ($(sqlq "$token"), $chain_id, $(sqlq "$symbol"), $decimals)
ON CONFLICT (token, chain_id) DO UPDATE
SET symbol = EXCLUDED.symbol, decimals = EXCLUDED.decimals;
SQL
    )
  else
    stmt=$(
      cat <<SQL
INSERT INTO ${TABLE} (token, chain_id, symbol, decimals)
VALUES ($(sqlq "$token"), $chain_id, $(sqlq "$symbol"), $decimals);
SQL
    )
  fi

  sql+="$stmt\n"
done
sql+="COMMIT;\n"

if ! printf "%b" "$sql" | grep -q "INSERT INTO"; then
  echo "Nothing to insert." >&2
  exit 0
fi

printf "%b" "$sql" | $PSQL -X -v ON_ERROR_STOP=1
