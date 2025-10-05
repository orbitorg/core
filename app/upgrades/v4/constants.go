package v4

import (
	icahosttypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/host/types"

	store "cosmossdk.io/store/types"

	"github.com/classic-terra/core/v3/app/upgrades"
)

const UpgradeName = "v4"

var Upgrade = upgrades.Upgrade{
	UpgradeName:          UpgradeName,
	CreateUpgradeHandler: CreateV4UpgradeHandler,
	StoreUpgrades:        store.StoreUpgrades{Added: []string{icahosttypes.StoreKey}},
}
