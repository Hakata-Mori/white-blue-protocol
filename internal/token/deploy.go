package token

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/white-blue-protocol/wblue/internal/storage"
	"github.com/white-blue-protocol/wblue/internal/types"
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

func Deploy(db *storage.DB, creator string, params *DeployParams, nonce uint64) (*types.BlueCoinConfig, error) {
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

	_, err := db.GetBlueCoinConfig(tokenID)
	if err == nil {
		return nil, fmt.Errorf("token already exists")
	}

	totalSupply := uint64(types.BlueCoinFixedSupply)
	poolAllocation := totalSupply * uint64(params.PoolRatio) / 100
	teamAllocation := totalSupply * uint64(params.TeamRatio) / 100

	now := time.Now().Unix()

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
		DeployedAt:     now,
	}

	state := &types.BlueCoinState{
		TokenID:        tokenID,
		TotalMinted:    totalSupply,
		PoolAllocation: poolAllocation,
		TeamLocked:     teamAllocation,
		TeamReleased:   0,
		LastUnlockTime: now,
	}

	if err := db.SaveBlueCoinConfig(config); err != nil {
		return nil, err
	}
	if err := db.SaveBlueCoinState(state); err != nil {
		return nil, err
	}

	pool := &types.AMMPool{
		TokenID:      tokenID,
		WhiteReserve: params.InitWhite,
		BlueReserve:  poolAllocation,
		K:            fmt.Sprintf("%d", params.InitWhite*poolAllocation),
		CreatedAt:    now,
	}
	if err := db.SavePool(pool); err != nil {
		return nil, err
	}

	return config, nil
}
