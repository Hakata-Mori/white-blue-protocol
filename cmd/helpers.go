package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/white-blue-protocol/wblue/internal/types"
)

func loadWalletByAddress(address string) (*types.KeyPair, error) {
	walletDir := filepath.Join(dataDir, "wallets")
	walletFile := filepath.Join(walletDir, address+".json")

	data, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %s", address)
	}

	var kp types.KeyPair
	if err := json.Unmarshal(data, &kp); err != nil {
		return nil, err
	}
	return &kp, nil
}
