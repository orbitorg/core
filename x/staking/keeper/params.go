package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetParams sets the x/staking module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	key := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	store := ctx.KVStore(key)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return k.getLegacyParams(ctx)
	}

	// if the byte array is not empty, we call the GetParams of the base keeper
	return k.Keeper.GetParams(ctx)
}

func (k Keeper) getLegacyParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.ss.GetParamSetIfExists(ctx, &params)
	return params
}
