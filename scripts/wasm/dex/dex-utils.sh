#!/bin/bash

source scripts/wasm/env-test-pre.sh

# Create asset info JSON (without amount)
create_asset_info_json() {
    local input=$1
    if [[ $input == terra* ]]; then
        echo "{\"token\":{\"contract_addr\":\"$input\"}}"
    else
        echo "{\"native_token\":{\"denom\":\"$input\"}}"
    fi
}

# Create full asset JSON with amount
create_asset_json() {
    local input=$1
    local amount=${2:-"0"} # Default amount to "0" if not provided
    if [[ $input == terra* ]]; then
        echo "{\"info\":{\"token\":{\"contract_addr\":\"$input\"}},\"amount\":\"$amount\"}"
    else
        echo "{\"info\":{\"native_token\":{\"denom\":\"$input\"}},\"amount\":\"$amount\"}"
    fi
}

# Function to create a pair
create_pair() {
    sleep $SLEEP_TIME 

    local factory_address=$1
    local token1=$2
    local token2=$3
    
    if [ -z "$token1" ] || [ -z "$token2" ]; then
        >&2 echo "Error: Both token addresses/denoms are required"
        return 1
    fi

    >&2 echo "Creating pair for tokens:"
    >&2 echo "Token 1: $token1"
    >&2 echo "Token 2: $token2"

    # Create asset JSON
    local asset1=$(create_asset_json "$token1")
    local asset2=$(create_asset_json "$token2")
    
    >&2 echo "Asset 1: $asset1"
    >&2 echo "Asset 2: $asset2"

    # Create pair message
    local msg=$(cat << EOF
{
    "create_pair": {
        "assets": [$asset1,$asset2]
    }
}
EOF
)

    # Execute create pair
    >&2 echo "Creating pair..."
    out=$($BINARY tx wasm execute "$factory_address" "$msg" \
        --from "$KEY" \
        --chain-id "$CHAIN_ID" \
        --gas 20000000 \
        --fees 1124975000uluna \
        --keyring-backend "$KEYRING" \
        --home "$HOME" \
        --output json \
        -y)
    
    sleep $SLEEP_TIME
    txhash=$(echo $out | jq -r '.txhash')
    
    # Query the tx and extract pair address
    sleep $SLEEP_TIME
    tx_response=$($BINARY q tx $txhash --output json)
    pair_address=$(echo "$tx_response" | jq -r '.logs[0].events[] | select(.type=="wasm").attributes[] | select(.key=="pair_contract_addr").value')
    
    printf "%s" "$pair_address"
}

# Function to query pair address from factory
query_pair_address() {
    local factory_address=$1
    local token1=$2
    local token2=$3

    local pair_query="{\"pair\":{\"asset_infos\":[$(create_asset_info_json $token1),$(create_asset_info_json $token2)]}}"
    local pair_info=$($BINARY query wasm contract-state smart $factory_address "$pair_query" --output json)
    echo $(echo $pair_info | jq -r '.data.contract_addr')
}

# Function to increase allowance for CW20 tokens
increase_allowance() {
    local token_address=$1
    local spender=$2
    local amount=$3

    >&2 echo "Increasing allowance for token $token_address..."
    out=$($BINARY tx wasm execute $token_address \
        "{\"increase_allowance\":{\"spender\":\"$spender\",\"amount\":\"$amount\"}}" \
        --from "$KEY" \
        --chain-id "$CHAIN_ID" \
        --gas 20000000 \
        --fees 1124975000uluna \
        --keyring-backend "$KEYRING" \
        --home "$HOME" \
        --output json \
        -y)
      
    
    sleep $SLEEP_TIME
}

# Function to provide liquidity
provide_liquidity() {
    local factory_address=$1
    local token1=$2
    local amount1=$3
    local token2=$4
    local amount2=$5

    >&2 echo "Providing liquidity..."
    >&2 echo "Token 1: $token1 Amount: $amount1"
    >&2 echo "Token 2: $token2 Amount: $amount2"

    # Query pair address
    local pair_address=$(query_pair_address "$factory_address" "$token1" "$token2")
    >&2 echo "Pair contract address: $pair_address"

    # Prepare assets
    local asset1=$(create_asset_json "$token1" "$amount1")
    local asset2=$(create_asset_json "$token2" "$amount2")

    # Handle CW20 tokens allowances
    if [[ $token1 == terra* ]]; then
        increase_allowance "$token1" "$pair_address" "$amount1"
    fi
    if [[ $token2 == terra* ]]; then
        increase_allowance "$token2" "$pair_address" "$amount2"
    fi

    # Prepare native token funds if needed
    local funds=""
    if [[ $token1 != terra* ]]; then
        funds="$funds--amount $amount1 "
    fi
    if [[ $token2 != terra* ]]; then
        funds="$funds--amount $amount2 "
    fi


    # Execute provide liquidity
    local msg=$(cat << EOF
{
    "provide_liquidity": {
        "assets": [$asset1,$asset2]
    }
}
EOF
)

    out=$($BINARY tx wasm execute "$pair_address" "$msg" \
        --from "$KEY" \
        --chain-id "$CHAIN_ID" \
        --gas 20000000 \
        --fees 1124975000uluna \
        $funds \
        --keyring-backend "$KEYRING" \
        --home "$HOME" \
        --output json \
        -y)

    sleep $SLEEP_TIME
    txhash=$(echo $out | jq -r '.txhash')
    
    # Query the tx and extract LP token amount
    sleep $SLEEP_TIME
    tx_response=$($BINARY q tx $txhash --output json)
}