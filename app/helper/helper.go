package helper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	oracleexported "github.com/classic-terra/core/v3/x/oracle/exported"
)

func IsOracleTx(msgs []sdk.Msg) bool {
	if len(msgs) == 0 {
		return false
	}
	for _, msg := range msgs {
		switch msg.(type) {
		case *oracleexported.MsgAggregateExchangeRatePrevote:
			continue
		case *oracleexported.MsgAggregateExchangeRateVote:
			continue
		default:
			return false
		}
	}

	return true
}
