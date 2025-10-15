package main

import (
	"os"

	core "github.com/classic-terra/core/v3/types"
	terraapp "github.com/classic-terra/core/v3/app"
	sdk "github.com/cosmos/cosmos-sdk/types"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
)

func init() {
	// Override the SDK's default bond denom to match Terra's configuration
	sdk.DefaultBondDenom = core.MicroLunaDenom
}

func main() {
	rootCmd, _ := NewRootCmd()

	if err := svrcmd.Execute(rootCmd, "", terraapp.DefaultNodeHome); err != nil {
		os.Exit(1)
	}
}
