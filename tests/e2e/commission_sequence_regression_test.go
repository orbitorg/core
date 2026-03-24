package e2e

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/classic-terra/core/v4/tests/e2e/configurer/chain"
	"github.com/classic-terra/core/v4/tests/e2e/initialization"
	tmconfig "github.com/cometbft/cometbft/config"
	"github.com/spf13/viper"
)

func isolateNodeFromPeers(node *chain.NodeConfig) (restore func(), err error) {
	cfgPath := filepath.Join(node.ConfigDir, "config", "config.toml")
	originalCfg, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	addrBookPath := filepath.Join(node.ConfigDir, "config", "addrbook.json")
	originalAddrBook, addrBookErr := os.ReadFile(addrBookPath)
	addrBookExists := addrBookErr == nil
	if addrBookErr != nil && !os.IsNotExist(addrBookErr) {
		return nil, addrBookErr
	}

	vpr := viper.New()
	vpr.SetConfigFile(cfgPath)
	if err := vpr.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := tmconfig.DefaultConfig()
	if err := vpr.Unmarshal(cfg); err != nil {
		return nil, err
	}

	cfg.P2P.PersistentPeers = ""
	cfg.P2P.Seeds = ""
	cfg.P2P.PexReactor = false
	tmconfig.WriteConfigFile(cfgPath, cfg)
	_ = os.Remove(addrBookPath)

	restore = func() {
		_ = os.WriteFile(cfgPath, originalCfg, 0o644)
		if addrBookExists {
			_ = os.WriteFile(addrBookPath, originalAddrBook, 0o644)
		} else {
			_ = os.Remove(addrBookPath)
		}
	}

	return restore, nil
}

func (s *IntegrationTestSuite) TestValidatorTxStuckInLocalMempoolPoisonsSequence() {
	chainCfg := s.configurer.GetChainConfig(0)
	observerNode := chainCfg.NodeConfigs[0]
	isolatedNode := chainCfg.NodeConfigs[1]

	restoreConfig, err := isolateNodeFromPeers(isolatedNode)
	s.Require().NoError(err)

	s.T().Cleanup(func() {
		_ = isolatedNode.Stop()
		restoreConfig()
		_ = isolatedNode.Run()
	})

	validatorAddr := isolatedNode.GetWallet(initialization.ValidatorWalletName)
	accountNumber, initialSequence, err := observerNode.QueryAccountInfo(validatorAddr)
	s.Require().NoError(err)

	initialDetails, err := observerNode.QueryValidatorDescriptionDetails(isolatedNode.OperatorAddress)
	s.Require().NoError(err)

	s.Require().NoError(isolatedNode.Stop())
	s.Require().NoError(isolatedNode.RunWithoutPeerWait())

	_, _, err = isolatedNode.GenerateSignAndBroadcastTxSync(
		[]string{
			"terrad", "tx", "staking", "edit-validator",
			"--details=local-mempool-sequence-poisoning",
			"--from=" + initialization.ValidatorWalletName,
		},
		initialization.ValidatorWalletName,
		accountNumber,
		initialSequence,
		"\"code\":0",
	)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		count, err := isolatedNode.QueryUnconfirmedTxCount()
		return err == nil && count >= 1
	}, initialization.OneMin, initialization.OneMin/60)

	chainCfg.WaitForNumHeights(2)

	sequenceAfterPendingEdit, err := observerNode.QueryAccountSequence(validatorAddr)
	s.Require().NoError(err)
	s.Require().Equal(initialSequence, sequenceAfterPendingEdit)

	detailsAfterPendingEdit, err := observerNode.QueryValidatorDescriptionDetails(isolatedNode.OperatorAddress)
	s.Require().NoError(err)
	s.Require().Equal(initialDetails, detailsAfterPendingEdit)

	out, errOut, err := isolatedNode.GenerateSignAndBroadcastTxSync(
		[]string{
			"terrad", "tx", "distribution", "withdraw-rewards", isolatedNode.OperatorAddress,
			"--commission",
			"--from=" + initialization.ValidatorWalletName,
		},
		initialization.ValidatorWalletName,
		accountNumber,
		initialSequence,
		"incorrect account sequence",
	)
	s.Require().NoError(err)

	combined := out + errOut
	s.Require().Contains(combined, "account sequence mismatch")
	s.Require().Contains(combined, "incorrect account sequence")
	s.Require().Contains(combined, "expected "+strconv.FormatUint(initialSequence+1, 10))
	s.Require().Contains(combined, "got "+strconv.FormatUint(initialSequence, 10))

	_, _, err = isolatedNode.GenerateSignAndBroadcastTxSync(
		[]string{
			"terrad", "tx", "distribution", "withdraw-rewards", isolatedNode.OperatorAddress,
			"--commission",
			"--from=" + initialization.ValidatorWalletName,
		},
		initialization.ValidatorWalletName,
		accountNumber,
		initialSequence+1,
		"\"code\":0",
	)
	s.Require().NoError(err)

	s.Require().Eventually(func() bool {
		count, err := isolatedNode.QueryUnconfirmedTxCount()
		return err == nil && count >= 2
	}, initialization.OneMin, initialization.OneMin/60)

	chainCfg.WaitForNumHeights(2)

	finalSequence, err := observerNode.QueryAccountSequence(validatorAddr)
	s.Require().NoError(err)
	s.Require().Equal(initialSequence, finalSequence)

	finalDetails, err := observerNode.QueryValidatorDescriptionDetails(isolatedNode.OperatorAddress)
	s.Require().NoError(err)
	s.Require().Equal(initialDetails, finalDetails)

	count, err := isolatedNode.QueryUnconfirmedTxCount()
	s.Require().NoError(err)
	s.Require().True(count >= 2, "expected at least 2 local unconfirmed txs, got %d", count)
}
