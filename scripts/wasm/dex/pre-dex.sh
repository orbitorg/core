source scripts/wasm/dex/fixture.sh
source scripts/wasm/dex/dex-utils.sh

echo "PRE-DEX: Adding pair"

PAIR_ADDRESS=$(create_pair $FACTORY_CONTRACT_ADDRESS $NATIVE_TOKEN $TOKEN_CONTRACT_ADDRESS)
echo "PAIR_ADDRESS: $PAIR_ADDRESS"

echo "PRE-DEX: Adding liquidity"
provide_liquidity $PAIR_ADDRESS $NATIVE_TOKEN '1000000000000uluna' $TOKEN_CONTRACT_ADDRESS "1000000000000$TOKEN_CONTRACT_ADDRESS"
