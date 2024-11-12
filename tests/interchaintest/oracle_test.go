package interchaintest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/classic-terra/core/v3/test/interchaintest/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/strangelove-ventures/interchaintest/v7"
	"github.com/strangelove-ventures/interchaintest/v7/chain/cosmos"
	"github.com/strangelove-ventures/interchaintest/v7/ibc"
	"github.com/strangelove-ventures/interchaintest/v7/testreporter"
	"github.com/strangelove-ventures/interchaintest/v7/testutil"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)


func TestOracle(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	numVals := 3
	numFullNodes := 3

	config, err := createConfig()
	require.NoError(t, err)

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "terra",
			ChainConfig:   config,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)

	terra := chains[0].(*cosmos.CosmosChain)

	ic := interchaintest.NewInterchain().AddChain(terra)

	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	ctx := context.Background()
	client, network := interchaintest.DockerSetup(t)

	err = ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: true,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = ic.Close()
	})

	err = testutil.WaitForBlocks(ctx, 1, terra)
	require.NoError(t, err)

	users := interchaintest.GetAndFundTestUsers(t, ctx, "default", genesisWalletAmount, terra, terra, terra, terra, terra, terra)

	err = testutil.WaitForBlocks(ctx, 5, terra)
	require.NoError(t, err)

	// Create error channels for both operations
	oracleErrorCh := make(chan error, len(terra.Validators))
	fundingErrorCh := make(chan error, len(terra.Validators))
	var wg sync.WaitGroup

	// Run oracle operations
	for i, val := range terra.Validators {
		wg.Add(2)
		go func(validator *cosmos.ChainNode, validatorIndex int) {
			defer wg.Done()

			// Seeding phase
			if err := helpers.ExecOracleMsgAggragatePrevote(ctx, validator, "salt", "1.123uusd"); err != nil {
				oracleErrorCh <- err
				return
			}

			// Wait for initial block
			if err := testutil.WaitForBlocks(ctx, 1, terra); err != nil {
				oracleErrorCh <- err
				return
			}

			// Oracle voting phase
			for i := 0; i < 2; i++ {
				if err := helpers.ExecOracleMsgAggragatePrevote(ctx, validator, "salt", "1.123uusd"); err != nil {
					oracleErrorCh <- err
					return
				}

				time.Sleep(500 * time.Millisecond)

				if err := helpers.ExecOracleMsgAggregateVote(ctx, validator, "salt", "1.123uusd"); err != nil {
					oracleErrorCh <- err
					return
				}

				if err := testutil.WaitForBlocks(ctx, 5, terra); err != nil {
					oracleErrorCh <- err
					return
				}
			}
		}(val, i)

	}

	for i, _ := range terra.Validators{
		wg.Add(1)
		go func(validatorIndex int) {
			defer wg.Done()

			for j := 0; j < 3; j++ {
				// First transfer
				err := terra.SendFunds(ctx, users[2*validatorIndex+1].KeyName(), ibc.WalletAmount{
					Address: string(users[2*validatorIndex].Address()),
					Denom:   terra.Config().Denom,
					Amount:  sdk.OneInt(),
				})

				if err != nil {
					fundingErrorCh <- err
					continue
				}



				// Second transfer
				err = terra.SendFunds(ctx, users[2*validatorIndex+1].KeyName(), ibc.WalletAmount{
					Address: string(users[2*validatorIndex].Address()),
					Denom:   terra.Config().Denom,
					Amount:  sdk.OneInt(),
				})

				if err != nil {
					fundingErrorCh <- err
					continue
				}
			}
		}(i)
	}


	// Wait for all goroutines to complete
	wg.Wait()
	close(oracleErrorCh)
	close(fundingErrorCh)

	// Check for any errors that occurred in oracle operations
	for err := range oracleErrorCh {
		require.NoError(t, err)
	}

	// Check for any errors that occurred in funding operations
	for err := range fundingErrorCh {
		require.NoError(t, err)
	}

	// Verify final validator state
	stdout, _, err := terra.Validators[0].ExecQuery(ctx, "staking", "validators")
	require.NoError(t, err)
	require.NotEmpty(t, stdout)

	terraValidators, _, err := helpers.UnmarshalValidators(*config.EncodingConfig, stdout)
	require.NoError(t, err)
	require.Equal(t, len(terraValidators), 3)
}
