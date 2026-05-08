package types

type AMMPool struct {
	TokenID        string `json:"tokenId"`
	WhiteReserve   uint64 `json:"whiteReserve"`
	BlueReserve    uint64 `json:"blueReserve"`
	K              string `json:"k"`
	TotalFeeBurned uint64 `json:"totalFeeBurned"`
	CreatedAt      int64  `json:"createdAt"`
	LastTradedAt   int64  `json:"lastTradedAt"`
}
