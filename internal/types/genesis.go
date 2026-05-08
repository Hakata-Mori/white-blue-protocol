package types

const (
	MaxWhiteSupply  = 1_000_000_000_000_000 // 1 billion * 1e6 micro
	BlockInterval   = 15                     // seconds
	BlocksPerYear   = 2_102_400              // 365*24*3600/15
	InitialReward   = 50_000_000             // 50 white coins in micro
	AnnualDecayRate = 10                     // 10% per year
	TxFee           = 1_000_000              // 1 white coin in micro (fixed fee)
	GenesisPremine  = 10_000_000_000         // 10,000 white coins for genesis validator
)

type GenesisConfig struct {
	ChainID          string            `json:"chainId"`
	GenesisValidator string            `json:"genesisValidator"`
	InitialBalances  map[string]uint64 `json:"initialBalances"`
}
