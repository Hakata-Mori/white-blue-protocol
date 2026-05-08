package token

import (
	"time"

	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/types"
)

const SecondsPerMonth = 30 * 24 * 3600

func ProcessVesting(db *storage.DB, blockTime int64) error {
	configs, err := db.ListBlueCoins()
	if err != nil {
		return err
	}

	for _, config := range configs {
		state, err := db.GetBlueCoinState(config.TokenID)
		if err != nil || state == nil {
			continue
		}

		if state.TeamLocked == 0 {
			continue
		}

		elapsed := blockTime - state.LastUnlockTime
		if elapsed < SecondsPerMonth {
			continue
		}

		months := elapsed / SecondsPerMonth
		for i := int64(0); i < months; i++ {
			if state.TeamLocked == 0 {
				break
			}

			unlock := config.ReleaseMonthly
			if unlock > state.TeamLocked {
				unlock = state.TeamLocked
			}

			state.TeamLocked -= unlock
			state.TeamReleased += unlock

			account := db.GetOrCreateAccount(config.MultiSigAddr)
			account.BlueBalances[config.TokenID] += unlock
			db.SaveAccount(account)
		}

		state.LastUnlockTime = blockTime - (elapsed % SecondsPerMonth)
		db.SaveBlueCoinState(state)
	}

	return nil
}

func GetNextUnlockTime(state *types.BlueCoinState) time.Time {
	return time.Unix(state.LastUnlockTime+SecondsPerMonth, 0)
}
