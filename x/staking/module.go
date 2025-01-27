package staking

import (
	customstakingkeeper "github.com/classic-terra/core/v3/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/cosmos/cosmos-sdk/x/staking"
)

// NewAppModule creates a new AppModule object
func NewAppModule(
	cdc codec.Codec,
	keeper *customstakingkeeper.Keeper,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	ls exported.Subspace,
) staking.AppModule {
	return staking.NewAppModule(cdc, &keeper.Keeper, ak, bk, ls)
}
