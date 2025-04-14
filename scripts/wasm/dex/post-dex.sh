source scripts/wasm/dex/fixture.sh
source scripts/wasm/dex/dex-utils.sh
source scripts/wasm/utils.sh

echo "POST-DEX: Asserting token balance"

TOKEN_BALANCE_BEFORE=$(get_token_balance $(get_address_from_key $KEY) $TOKEN_CONTRACT_ADDRESS)
echo "TOKEN_BALANCE_BEFORE: $TOKEN_BALANCE_BEFORE"

TOKEN_BALANCE_AFTER=$(get_token_balance $(get_address_from_key $KEY) $TOKEN_CONTRACT_ADDRESS)
echo "TOKEN_BALANCE_AFTER: $TOKEN_BALANCE_AFTER"

# PAIR_ADDRESS=$(create_pair $FACTORY_CONTRACT_ADDRESS $NATIVE_TOKEN $TOKEN_CONTRACT_ADDRESS)
# echo "PAIR_ADDRESS: $PAIR_ADDRESS"

# echo "POST-DEX: Adding liquidity"
# provide_liquidity $PAIR_ADDRESS $NATIVE_TOKEN '10000000000' $TOKEN_CONTRACT_ADDRESS "1000000000000"
