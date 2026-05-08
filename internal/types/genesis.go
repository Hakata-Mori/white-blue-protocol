package types

const (
	MaxWhiteSupply  = 1_000_000_000_000_000 // 1 billion * 1e6 micro
	BlockInterval   = 15                     // seconds
	BlocksPerYear   = 2_102_400              // 365*24*3600/15
	InitialReward   = 50_000_000             // 50 white coins in micro
	AnnualDecayRate = 10                     // 10% per year
	FeeRate         = 1000                   // 0.1% = 1/1000
	MinFee          = 1_000                  // 0.001 WC minimum fee in micro
	GenesisPremine  = 10_000_000_000         // 10,000 white coins for genesis validator
)

func CalcFee(amount uint64) uint64 {
	fee := amount / FeeRate
	if fee < MinFee {
		return MinFee
	}
	return fee
}

type GenesisConfig struct {
	ChainID          string            `json:"chainId"`
	GenesisValidator string            `json:"genesisValidator"`
	InitialBalances  map[string]uint64 `json:"initialBalances"`
}
