#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "Usage: ./gen.sh <ContractPath:ContractName> (e.g., Custody.sol:Custody)"
  exit 1
fi

INPUT="$1"
IFS=':' read -r SOLFILE CONTRACT <<<"$INPUT"

# Paths
JSON_PATH="../../contract/out/$SOLFILE/$CONTRACT.json"
COMBINED_JSON="combined.json"
OUTPUT_GO="bindings.go"
PKG_NAME="nitrolite"

# Validate file exists
if [ ! -f "$JSON_PATH" ]; then
  echo "Error: Artifact not found at $JSON_PATH"
  exit 1
fi

# Extract combined-json
jq --arg name "$SOLFILE:$CONTRACT" '{contracts: { ($name): { abi: .abi, bin: .bytecode.object }}}' "$JSON_PATH" >"$COMBINED_JSON"

# Generate Go bindings
go run github.com/ethereum/go-ethereum/cmd/abigen@latest \
  --combined-json "$COMBINED_JSON" \
  --pkg "$PKG_NAME" \
  --out "$OUTPUT_GO"

echo "Generated Go bindings at $OUTPUT_GO"
