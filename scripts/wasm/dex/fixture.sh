#!/bin/bash
source scripts/wasm/env-test-pre.sh
source scripts/wasm/helpers.sh

# Configuration
FACTORY_CONTRACT_PATH="scripts/wasm/dex/artifacts/terraswap_factory.wasm"
PAIR_CONTRACT_PATH="scripts/wasm/dex/artifacts/terraswap_pair.wasm"
ROUTER_CONTRACT_PATH="scripts/wasm/dex/artifacts/terraswap_router.wasm"
TOKEN_CONTRACT_PATH="scripts/wasm/dex/artifacts/terraswap_token.wasm"

TREASURY="terra1nnj62ced7cpk2ll0cpavwqv9fufqfgznuwk4nm"

# Token configuration
TOKEN_NAME="Test Token"
TOKEN_SYMBOL="TEST"
TOKEN_DECIMALS=6

# Initialize variables to store code IDs and addresses
FACTORY_CODE_ID=""
PAIR_CODE_ID=""
ROUTER_CODE_ID=""
TOKEN_CODE_ID=""

FACTORY_CONTRACT_ADDRESS=""
ROUTER_CONTRACT_ADDRESS=""
TOKEN_CONTRACT_ADDRESS=""

# Check if files exist before uploading
for contract in "$FACTORY_CONTRACT_PATH" "$PAIR_CONTRACT_PATH" "$ROUTER_CONTRACT_PATH" "$TOKEN_CONTRACT_PATH"; do
    if [ ! -f "$contract" ]; then
        echo "Error: Contract file $contract not found!"
        exit 1
    fi
done

# uploading the contracts
FACTORY_CODE_ID=$(upload_contract "$FACTORY_CONTRACT_PATH")
PAIR_CODE_ID=$(upload_contract "$PAIR_CONTRACT_PATH")
ROUTER_CODE_ID=$(upload_contract "$ROUTER_CONTRACT_PATH")
TOKEN_CODE_ID=$(upload_contract "$TOKEN_CONTRACT_PATH")

echo "Uploaded contracts with code IDs:"
echo "FACTORY_CODE_ID: $FACTORY_CODE_ID"
echo "PAIR_CODE_ID: $PAIR_CODE_ID"
echo "ROUTER_CODE_ID: $ROUTER_CODE_ID"
echo "TOKEN_CODE_ID: $TOKEN_CODE_ID"

# Create token instantiation message
echo "Instantiating token contract..."
TOKEN_MSG=$(cat << EOF
{
    "name": "$TOKEN_NAME",
    "symbol": "$TOKEN_SYMBOL",
    "decimals": $TOKEN_DECIMALS,
    "initial_balances": [{
        "address": "$TREASURY",
        "amount": "1000000000000000"
    }],
    "mint": {
        "minter": "$TREASURY",
        "cap": "1000000000000000"
    },
    "marketing": null
}
EOF
)

# Instantiate token contract
TOKEN_CONTRACT_ADDRESS=$(instantiate_contract "$TOKEN_CONTRACT_PATH" "$TOKEN_CODE_ID" "$TOKEN_MSG" "$TOKEN_NAME")
echo "Token contract address: $TOKEN_CONTRACT_ADDRESS"

# Create factory instantiation message
echo "Instantiating factory contract..."
FACTORY_MSG=$(cat << EOF
{
  "token_code_id": $TOKEN_CODE_ID,
  "pair_code_id": $PAIR_CODE_ID,
  "platform_treasury": "$TREASURY",
  "dev_treasury": "$TREASURY"
}
EOF
)

# Instantiate factory contract
FACTORY_CONTRACT_ADDRESS=$(instantiate_contract "$FACTORY_CONTRACT_PATH" "$FACTORY_CODE_ID" "$FACTORY_MSG" "LUNC Terraswap Factory")
echo "Factory contract address: $FACTORY_CONTRACT_ADDRESS"

# Create router instantiation message
ROUTER_MSG=$(cat << EOF
{
  "terraswap_factory": "$FACTORY_CONTRACT_ADDRESS"
}
EOF
)

# Instantiate router contract
ROUTER_CONTRACT_ADDRESS=$(instantiate_contract "$ROUTER_CONTRACT_PATH" "$ROUTER_CODE_ID" "$ROUTER_MSG" "LUNC Terraswap Router")
echo "Router contract address: $ROUTER_CONTRACT_ADDRESS"

# Save all contract addresses and code IDs
cat > scripts/wasm/dex/deployment_info.txt << EOL
TOKEN_CODE_ID=$TOKEN_CODE_ID
PAIR_CODE_ID=$PAIR_CODE_ID
FACTORY_CODE_ID=$FACTORY_CODE_ID
ROUTER_CODE_ID=$ROUTER_CODE_ID
TOKEN_CONTRACT_ADDRESS=$TOKEN_CONTRACT_ADDRESS
FACTORY_CONTRACT_ADDRESS=$FACTORY_CONTRACT_ADDRESS
ROUTER_CONTRACT_ADDRESS=$ROUTER_CONTRACT_ADDRESS
EOL

echo "----------------------------------------"
echo "Deployment Summary:"
echo "Token Code ID: $TOKEN_CODE_ID"
echo "Token Address: $TOKEN_CONTRACT_ADDRESS"
echo "Factory Code ID: $FACTORY_CODE_ID"
echo "Factory Address: $FACTORY_CONTRACT_ADDRESS"
echo "Router Code ID: $ROUTER_CODE_ID"
echo "Router Address: $ROUTER_CONTRACT_ADDRESS"
echo "----------------------------------------"
echo "All information has been saved to scripts/wasm/dex/deployment_info.txt"


echo "----------------------------------------"
echo "Setting up the config"
echo "CONFIG: Adding the native token decimals"

msg="{\"add_native_token_decimals\":{\"denom\":\"uluna\",\"decimals\":6}}"

out=$($BINARY tx wasm execute $FACTORY_CONTRACT_ADDRESS \
    "$msg" \
    --from $KEY \
    --chain-id $CHAIN_ID \
    --gas $GAS \
    --fees 1124975000uluna \
    --amount 1000000uluna \
    --broadcast-mode sync \
    --keyring-backend $KEYRING \
    --home $HOME \
    --output json \
    -y)