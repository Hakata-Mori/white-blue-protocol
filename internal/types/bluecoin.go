package types

type BlueCoinConfig struct {
	TokenID        string   `json:"tokenId"`
	Name           string   `json:"name"`
	Symbol         string   `json:"symbol"`
	Creator        string   `json:"creator"`
	TotalSupply    uint64   `json:"totalSupply"`
	PoolRatio      uint8    `json:"poolRatio"`
	TeamRatio      uint8    `json:"teamRatio"`
	InitWhite      uint64   `json:"initWhite"`
	ReleaseMonthly uint64   `json:"releaseMonthly"`
	MultiSigAddr   string   `json:"multiSigAddr"`
	SourceURLs     []string `json:"sourceUrls"`
	DeployedAt     int64    `json:"deployedAt"`
	DeployTxHash   string   `json:"deployTxHash"`
}

type BlueCoinState struct {
	TokenID         string `json:"tokenId"`
	TotalMinted     uint64 `json:"totalMinted"`
	PoolAllocation  uint64 `json:"poolAllocation"`
	TeamLocked      uint64 `json:"teamLocked"`
	TeamReleased    uint64 `json:"teamReleased"`
	LastUnlockTime  int64  `json:"lastUnlockTime"`
	Burned          uint64 `json:"burned"`
}

const BlueCoinFixedSupply = 1_000_000_000_000 // 1,000,000 * 1e6 micro
