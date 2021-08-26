#!/bin/bash

set -o xtrace
curl -s http://localhost:8545 -H 'Content-Type: application/json' -d '{"method":"'$1'","id":1,"jsonrpc":"2.0", "params":['$2']}' | jq '.["result"]' > data.json 2>&1
cat openrpc.json | jq '.["methods"][] | select(.name == "'$1'") | .["result"]["schema"]' > schema.json
ajv validate -s schema.json -d data.json
# rm schema.json data.json
