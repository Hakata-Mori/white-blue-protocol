package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"syscall"

	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/keystore"
	"github.com/white-blue-protocol/wblue/internal/types"
	"golang.org/x/term"
)

func loadWalletByAddress(address string) (*types.KeyPair, error) {
	walletDir := filepath.Join(dataDir, "wallets")
	walletFile := filepath.Join(walletDir, address+".json")

	data, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, fmt.Errorf("wallet not found: %s", address)
	}

	if keystore.IsKeystoreFile(data) {
		ks, err := keystore.Load(walletFile)
		if err != nil {
			return nil, err
		}

		password, err := readPassword("Enter wallet password: ")
		if err != nil {
			return nil, err
		}

		privKey, err := keystore.Decrypt(ks, password)
		if err != nil {
			return nil, err
		}

		return &types.KeyPair{
			PrivateKey: privKey,
			PublicKey:  ks.PublicKey,
			Address:    ks.Address,
		}, nil
	}

	var kp types.KeyPair
	if err := json.Unmarshal(data, &kp); err != nil {
		return nil, err
	}
	return &kp, nil
}

func readPassword(prompt string) (string, error) {
	if valPassword != "" {
		return valPassword, nil
	}
	if envPw := os.Getenv("WBLUE_VALIDATOR_PASSWORD"); envPw != "" {
		return envPw, nil
	}
	fmt.Print(prompt)
	if term.IsTerminal(int(syscall.Stdin)) {
		pw, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return "", err
		}
		return string(pw), nil
	}
	var pw string
	if _, err := fmt.Scanln(&pw); err != nil {
		return "", err
	}
	return pw, nil
}

func getAccountNonce(address string) (uint64, error) {
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/wallet/%s", apiURL, address))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var account types.Account
	json.NewDecoder(resp.Body).Decode(&account)
	return account.Nonce, nil
}

func signAndSubmit(tx *types.Transaction, privKey string) error {
	txCopy := *tx
	txCopy.Signature = ""
	txCopy.Hash = ""
	txData, _ := json.Marshal(txCopy)
	tx.Hash = crypto.SHA256Hex(txData)

	sig, err := crypto.Sign(privKey, txData)
	if err != nil {
		return fmt.Errorf("sign: %w", err)
	}
	tx.Signature = sig

	body, _ := json.Marshal(tx)
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/tx/submit", apiURL), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		msg, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("submit failed: %s", string(msg))
	}
	return nil
}
