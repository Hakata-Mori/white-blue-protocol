package token

import (
	"github.com/white-blue-protocol/wblue/internal/storage"
	bolt "go.etcd.io/bbolt"
)

const SecondsPerMonth = 30 * 24 * 3600

func ProcessVesting(db *storage.DB, blockTime int64) error {
	return db.Update(func(btx *bolt.Tx) error {
		return ProcessVestingInTx(btx, blockTime)
	})
}

func ProcessVestingInTx(btx *bolt.Tx, blockTime int64) error {
	configs, err := storage.ListBlueCoinsInTx(btx)
	if err != nil {
		return err
	}

	for _, config := range configs {
		state, err := storage.GetBlueCoinStateInTx(btx, config.TokenID)
		if err != nil {
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

			account := storage.GetOrCreateAccountInTx(btx, config.MultiSigAddr)
			account.BlueBalances[config.TokenID] += unlock
			if err := storage.SaveAccountInTx(btx, account); err != nil {
				return err
			}
		}

		state.LastUnlockTime = blockTime - (elapsed % SecondsPerMonth)
		if err := storage.SaveBlueCoinStateInTx(btx, state); err != nil {
			return err
		}
	}

	return nil
}
