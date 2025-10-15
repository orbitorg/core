package main

import (
	"os"

	terraapp "github.com/classic-terra/core/v3/app"
	core "github.com/classic-terra/core/v3/types"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
