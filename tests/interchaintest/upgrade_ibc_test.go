package interchaintest

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/interchaintest/v10"
	"github.com/cosmos/interchaintest/v10/chain/cosmos"
	"github.com/cosmos/interchaintest/v10/conformance"
	"github.com/cosmos/interchaintest/v10/ibc"
	"github.com/cosmos/interchaintest/v10/testreporter"
	"github.com/cosmos/interchaintest/v10/testutil"
	"github.com/icza/dyno"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	haltHeightDelta    = 10
	blocksAfterUpgrade = 10
	shortVotingPeriod  = "10s"
	shortDepositPeriod = "10s"
)

// TestTerraUpgradeIBC spins up Terra and Gaia, performs an on-chain software-upgrade
// for Terra at a future height, upgrades the container image (same image is fine when
// the handler already exists), and verifies IBC conformance before and after.
func TestTerraUpgradeIBC(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}
	t.Parallel()

	// single validator, no full nodes for speed/stability
	numVals := 1
	numFullNodes := 0

	// Terra config with short governance periods for faster proposal processing
	terraCfg, err := createConfig()
	require.NoError(t, err)
	terraCfg.ModifyGenesis = shortGovModifyGenesis()

	cf := interchaintest.NewBuiltinChainFactory(zaptest.NewLogger(t), []*interchaintest.ChainSpec{
		{
			Name:          "terra",
			ChainConfig:   terraCfg,
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
		{
			Name:          "gaia",
			// Keep consistent with other suite tests for compatibility
			Version:       "v12.0.0",
			NumValidators: &numVals,
			NumFullNodes:  &numFullNodes,
		},
	})

	chains, err := cf.Chains(t.Name())
	require.NoError(t, err)
	terra, gaia := chains[0].(*cosmos.CosmosChain), chains[1].(*cosmos.CosmosChain)

	client, network := interchaintest.DockerSetup(t)

	// Use go-relayer for consistency with this suite
	r := interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)).
		Build(t, client, network)

	const (
		path        = "terra-gaia-upgrade-test"
		relayerName = "relayer"
	)

	ic := interchaintest.NewInterchain().
		AddChain(terra).
		AddChain(gaia).
		AddRelayer(r, relayerName).
		AddLink(interchaintest.InterchainLink{
			Chain1:  terra,
			Chain2:  gaia,
			Relayer: r,
			Path:    path,
		})

	ctx := context.Background()
	rep := testreporter.NewNopReporter()
	eRep := rep.RelayerExecReporter(t)

	require.NoError(t, ic.Build(ctx, eRep, interchaintest.InterchainBuildOptions{
		TestName:         t.Name(),
		Client:           client,
		NetworkID:        network,
		SkipPathCreation: false,
	}))
	t.Cleanup(func() { _ = ic.Close() })

	// Fund a user on Terra for proposal deposit and fees
	userFunds := math.NewInt(10_000_000_000)
	users := interchaintest.GetAndFundTestUsers(t, ctx, t.Name(), userFunds, terra)
	terraUser := users[0]

	// IBC conformance before upgrade
	conformance.TestChainPair(t, ctx, client, network, terra, gaia, interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)), rep, r, path)

	// Plan and submit upgrade
	height, err := terra.Height(ctx)
	require.NoError(t, err, "error fetching height before submit upgrade proposal")
	haltHeight := height + haltHeightDelta

	// Choose an upgrade name that exists in this binary; "v14" is present in app/upgrades.
	const upgradeName = "v14"

	prop := cosmos.SoftwareUpgradeProposal{
		Deposit:     "500000000" + terra.Config().Denom, // greater than min deposit
		Title:       "Terra Chain Upgrade",
		Name:        upgradeName,
		Description: "Interchaintest software upgrade",
		Height:      haltHeight,
	}
	upgradeTx, err := terra.UpgradeProposal(ctx, terraUser.KeyName(), prop)
	require.NoError(t, err, "error submitting software upgrade proposal tx")

	propID, err := strconv.ParseUint(upgradeTx.ProposalID, 10, 64)
	require.NoError(t, err, "failed to convert proposal ID to uint64")

	err = terra.VoteOnProposalAllValidators(ctx, propID, cosmos.ProposalVoteYes)
	require.NoError(t, err, "failed to submit votes")

	_, err = cosmos.PollForProposalStatus(ctx, terra, height, height+haltHeightDelta, propID, govv1beta1.StatusPassed)
	require.NoError(t, err, "proposal status did not change to passed in expected number of blocks")

	height, err = terra.Height(ctx)
	require.NoError(t, err, "error fetching height before upgrade")

	timeoutCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	// Expect timeout due to chain halt at upgrade height.
	_ = testutil.WaitForBlocks(timeoutCtx, int(haltHeight-height)+1, terra)

	height, err = terra.Height(ctx)
	require.NoError(t, err, "error fetching height after chain should have halted")
	require.Equal(t, haltHeight, height, "height is not equal to halt height")

	// Stop nodes, "upgrade" image (same repo/version is sufficient if handler exists), and restart.
	err = terra.StopAllNodes(ctx)
	require.NoError(t, err, "error stopping node(s)")

	// Use the configured docker image from CI or local build.
	repo, version := GetDockerImageInfo()
	terra.UpgradeVersion(ctx, client, repo, version)

	err = terra.StartAllNodes(ctx)
	require.NoError(t, err, "error starting upgraded node(s)")

	timeoutCtx, cancel = context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	err = testutil.WaitForBlocks(timeoutCtx, blocksAfterUpgrade, terra)
	require.NoError(t, err, "chain did not produce blocks after upgrade")

	// IBC conformance after upgrade on the same path
	conformance.TestChainPair(t, ctx, client, network, terra, gaia, interchaintest.NewBuiltinRelayerFactory(ibc.CosmosRly, zaptest.NewLogger(t)), rep, r, path)
}

// shortGovModifyGenesis returns a ModifyGenesis function that sets short governance
// voting and deposit periods for faster software-upgrade proposals on Terra.
func shortGovModifyGenesis() func(ibc.ChainConfig, []byte) ([]byte, error) {
	base := ModifyGenesis()
	return func(chainConfig ibc.ChainConfig, genbz []byte) ([]byte, error) {
		out, err := base(chainConfig, genbz)
		if err != nil {
			return nil, err
		}
		// apply overrides for faster governance in this test
		g := map[string]interface{}{}
		if err := jsonUnmarshal(out, &g); err != nil {
			return nil, err
		}
		if err := dyno.Set(g, shortVotingPeriod, "app_state", "gov", "params", "voting_period"); err != nil {
			return nil, err
		}
		if err := dyno.Set(g, shortDepositPeriod, "app_state", "gov", "params", "max_deposit_period"); err != nil {
			return nil, err
		}
		return jsonMarshal(g)
	}
}

// jsonUnmarshal/jsonMarshal small wrappers to keep imports local to this file.
func jsonUnmarshal(bz []byte, v any) error {
	return json.Unmarshal(bz, v)
}
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}


