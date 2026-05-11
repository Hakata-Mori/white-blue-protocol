package token

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/types"
	bolt "go.etcd.io/bbolt"
)

type DeployParams struct {
	Name           string   `json:"name"`
	Symbol         string   `json:"symbol"`
	PoolRatio      uint8    `json:"poolRatio"`
	TeamRatio      uint8    `json:"teamRatio"`
	InitWhite      uint64   `json:"initWhite"`
	ReleaseMonthly uint64   `json:"releaseMonthly"`
	MultiSigAddr   string   `json:"multiSigAddr"`
	SourceURLs     []string `json:"sourceUrls"`
}

func GenerateTokenID(creator, name string, nonce uint64) string {
	input := fmt.Sprintf("%s:%s:%d", creator, name, nonce)
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:8])
}

func Deploy(db *storage.DB, creator string, params *DeployParams, nonce uint64, blockTime int64) (*types.BlueCoinConfig, error) {
	var config *types.BlueCoinConfig
	err := db.Update(func(btx *bolt.Tx) error {
		var e error
		config, e = DeployInTx(btx, creator, params, nonce, blockTime)
		return e
	})
	return config, err
}

func DeployInTx(btx *bolt.Tx, creator string, params *DeployParams, nonce uint64, blockTime int64) (*types.BlueCoinConfig, error) {
	if params.PoolRatio+params.TeamRatio != 100 {
		return nil, fmt.Errorf("poolRatio + teamRatio must equal 100")
	}
	if params.Name == "" || params.Symbol == "" {
		return nil, fmt.Errorf("name and symbol are required")
	}
	if params.InitWhite == 0 {
		return nil, fmt.Errorf("initWhite must be > 0")
	}

	tokenID := GenerateTokenID(creator, params.Name, nonce)

	_, err := storage.GetBlueCoinConfigInTx(btx, tokenID)
	if err == nil {
		return nil, fmt.Errorf("token already exists")
	}

	totalSupply := uint64(types.BlueCoinFixedSupply)
	poolAllocation := totalSupply * uint64(params.PoolRatio) / 100
	teamAllocation := totalSupply * uint64(params.TeamRatio) / 100

	config := &types.BlueCoinConfig{
		TokenID:        tokenID,
		Name:           params.Name,
		Symbol:         params.Symbol,
		Creator:        creator,
		TotalSupply:    totalSupply,
		PoolRatio:      params.PoolRatio,
		TeamRatio:      params.TeamRatio,
		InitWhite:      params.InitWhite,
		ReleaseMonthly: params.ReleaseMonthly,
		MultiSigAddr:   params.MultiSigAddr,
		SourceURLs:     params.SourceURLs,
		DeployedAt:     blockTime,
	}

	state := &types.BlueCoinState{
		TokenID:        tokenID,
		TotalMinted:    totalSupply,
		PoolAllocation: poolAllocation,
		TeamLocked:     teamAllocation,
		TeamReleased:   0,
		LastUnlockTime: blockTime,
	}

	if err := storage.SaveBlueCoinConfigInTx(btx, config); err != nil {
		return nil, err
	}
	if err := storage.SaveBlueCoinStateInTx(btx, state); err != nil {
		return nil, err
	}

	k := new(big.Int).Mul(
		new(big.Int).SetUint64(params.InitWhite),
		new(big.Int).SetUint64(poolAllocation),
	)

	pool := &types.AMMPool{
		TokenID:      tokenID,
		WhiteReserve: params.InitWhite,
		BlueReserve:  poolAllocation,
		K:            k.String(),
		CreatedAt:    blockTime,
	}
	if err := storage.SavePoolInTx(btx, pool); err != nil {
		return nil, err
	}

	return config, nil
}
