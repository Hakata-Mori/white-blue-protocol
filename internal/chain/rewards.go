package chain

import (
	"github.com/white-blue-protocol/wblue/internal/types"
)

func CalcReward(height uint64, totalMinted uint64) uint64 {
	if totalMinted >= types.MaxWhiteSupply {
		return 0
	}

	year := height / types.BlocksPerYear
	reward := uint64(types.InitialReward)
	for i := uint64(0); i < year; i++ {
		reward = reward * 9 / 10
	}

	remaining := types.MaxWhiteSupply - totalMinted
	if reward > remaining {
		return remaining
	}
	return reward
}
