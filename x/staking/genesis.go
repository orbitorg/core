package staking

import (
	customstakingkeeper "github.com/classic-terra/core/v3/x/staking/keeper"

	tmtypes "github.com/cometbft/cometbft/types"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// WriteValidators returns a slice of bonded genesis validators.
func WriteValidators(ctx sdk.Context, keeper *customstakingkeeper.Keeper) (vals []tmtypes.GenesisValidator, returnErr error) {
	keeper.IterateLastValidators(ctx, func(_ int64, validator types.ValidatorI) (stop bool) {
		pk, err := validator.ConsPubKey()
		if err != nil {
			returnErr = err
			return true
		}
		tmPk, err := cryptocodec.ToTmPubKeyInterface(pk)
		if err != nil {
			returnErr = err
			return true
		}

		vals = append(vals, tmtypes.GenesisValidator{
			Address: sdk.ConsAddress(tmPk.Address()).Bytes(),
			PubKey:  tmPk,
			Power:   validator.GetConsensusPower(keeper.PowerReduction(ctx)),
			Name:    validator.GetMoniker(),
		})

		return false
	})

	return
}
