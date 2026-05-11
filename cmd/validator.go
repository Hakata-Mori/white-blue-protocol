package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/white-blue-protocol/wblue/internal/crypto"
	"github.com/white-blue-protocol/wblue/internal/types"
)

var validatorCmd = &cobra.Command{
	Use:   "validator",
	Short: "Validator operations",
}

var validatorJoinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join as a validator (requires 24h online + PoW)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		if fromAddr == "" {
			return fmt.Errorf("--from is required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/chain/status", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()
		var status struct {
			Height uint64 `json:"height"`
		}
		json.NewDecoder(resp.Body).Decode(&status)

		resp2, err := http.Get(fmt.Sprintf("%s/api/v1/wallet/%s", apiURL, kp.Address))
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		var account types.Account
		json.NewDecoder(resp2.Body).Decode(&account)

		fmt.Println("Computing PoW nonce (this may take a few seconds)...")
		nonce := solvePoW(kp.Address, status.Height+1)
		fmt.Printf("PoW solved: nonce=%d\n", nonce)

		payload, _ := json.Marshal(struct {
			Nonce uint64 `json:"nonce"`
		}{Nonce: nonce})

		tx := types.Transaction{
			Type:      types.TxValidatorJoin,
			From:      kp.Address,
			Amount:    status.Height + 1,
			Nonce:     account.Nonce + 1,
			PublicKey: kp.PublicKey,
			Payload:   payload,
			Timestamp: time.Now().Unix(),
		}

		txCopy := tx
		txCopy.Signature = ""
		txCopy.Hash = ""
		txData, _ := json.Marshal(txCopy)
		tx.Hash = crypto.SHA256Hex(txData)

		sig, err := crypto.Sign(kp.PrivateKey, txData)
		if err != nil {
			return err
		}
		tx.Signature = sig

		body, _ := json.Marshal(tx)
		submitResp, err := http.Post(fmt.Sprintf("%s/api/v1/tx/submit", apiURL), "application/json",
			bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer submitResp.Body.Close()

		if submitResp.StatusCode != 200 {
			var errMsg string
			buf := make([]byte, 1024)
			n, _ := submitResp.Body.Read(buf)
			errMsg = string(buf[:n])
			return fmt.Errorf("submit failed: %s", errMsg)
		}

		fmt.Printf("Validator join submitted: %s\n", tx.Hash[:16])
		fmt.Println("(Will be confirmed in next block)")
		return nil
	},
}

var validatorExitCmd = &cobra.Command{
	Use:   "exit",
	Short: "Exit as a validator (stake will be burned)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		if fromAddr == "" {
			return fmt.Errorf("--from is required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/chain/status", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()
		var status struct {
			Height uint64 `json:"height"`
		}
		json.NewDecoder(resp.Body).Decode(&status)

		resp2, err := http.Get(fmt.Sprintf("%s/api/v1/wallet/%s", apiURL, kp.Address))
		if err != nil {
			return err
		}
		defer resp2.Body.Close()
		var account types.Account
		json.NewDecoder(resp2.Body).Decode(&account)

		tx := types.Transaction{
			Type:      types.TxValidatorExit,
			From:      kp.Address,
			Amount:    status.Height + 1,
			Nonce:     account.Nonce + 1,
			PublicKey: kp.PublicKey,
			Timestamp: time.Now().Unix(),
		}

		txCopy := tx
		txCopy.Signature = ""
		txCopy.Hash = ""
		txData, _ := json.Marshal(txCopy)
		tx.Hash = crypto.SHA256Hex(txData)

		sig, err := crypto.Sign(kp.PrivateKey, txData)
		if err != nil {
			return err
		}
		tx.Signature = sig

		body, _ := json.Marshal(tx)
		submitResp, err := http.Post(fmt.Sprintf("%s/api/v1/tx/submit", apiURL), "application/json",
			bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer submitResp.Body.Close()

		if submitResp.StatusCode != 200 {
			var errMsg string
			buf := make([]byte, 1024)
			n, _ := submitResp.Body.Read(buf)
			errMsg = string(buf[:n])
			return fmt.Errorf("submit failed: %s", errMsg)
		}

		fmt.Printf("Validator exit submitted: %s\n", tx.Hash[:16])
		fmt.Println("Stake will be burned. (Confirmed in next block)")
		return nil
	},
}

var validatorStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show validator set status",
	RunE: func(cmd *cobra.Command, args []string) error {
		resp, err := http.Get(fmt.Sprintf("%s/api/v1/validators", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()

		var vs types.ValidatorSet
		if err := json.NewDecoder(resp.Body).Decode(&vs); err != nil {
			return err
		}

		fmt.Printf("Validators (%d total):\n", len(vs.Validators))
		for _, v := range vs.Validators {
			addr := v.Address
			if len(addr) > 14 {
				addr = addr[:14] + "..."
			}
			fmt.Printf("  %s  status=%-10s joined=#%d  lastHB=#%d\n",
				addr, v.Status, v.JoinHeight, v.LastHeartbeatHeight)
		}
		return nil
	},
}

var validatorHeartbeatCmd = &cobra.Command{
	Use:   "heartbeat",
	Short: "Send a heartbeat to prove uptime (for candidates)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fromAddr, _ := cmd.Flags().GetString("from")
		if fromAddr == "" {
			return fmt.Errorf("--from is required")
		}

		kp, err := loadWalletByAddress(fromAddr)
		if err != nil {
			return err
		}

		resp, err := http.Get(fmt.Sprintf("%s/api/v1/chain/status", apiURL))
		if err != nil {
			return fmt.Errorf("node not running? %w", err)
		}
		defer resp.Body.Close()
		var status struct {
			Height uint64 `json:"height"`
		}
		json.NewDecoder(resp.Body).Decode(&status)

		nonce, err := getAccountNonce(kp.Address)
		if err != nil {
			return err
		}

		tx := types.Transaction{
			Type:      types.TxHeartbeat,
			From:      kp.Address,
			Amount:    status.Height + 1,
			Nonce:     nonce + 1,
			PublicKey: kp.PublicKey,
			Timestamp: time.Now().Unix(),
		}

		if err := signAndSubmit(&tx, kp.PrivateKey); err != nil {
			return err
		}

		fmt.Printf("Heartbeat sent at height %d\n", status.Height+1)
		return nil
	},
}

func solvePoW(address string, height uint64) uint64 {
	requiredBytes := types.GetPoWDifficulty() / 8
	for nonce := uint64(0); ; nonce++ {
		data := fmt.Sprintf("%s:%d:%d", address, height, nonce)
		hash := sha256.Sum256([]byte(data))
		allZero := true
		for i := 0; i < requiredBytes; i++ {
			if hash[i] != 0 {
				allZero = false
				break
			}
		}
		if allZero {
			return nonce
		}
	}
}

func init() {
	validatorJoinCmd.Flags().String("from", "", "Validator address")
	validatorExitCmd.Flags().String("from", "", "Validator address")
	validatorHeartbeatCmd.Flags().String("from", "", "Candidate address")
	validatorCmd.AddCommand(validatorJoinCmd)
	validatorCmd.AddCommand(validatorExitCmd)
	validatorCmd.AddCommand(validatorHeartbeatCmd)
	validatorCmd.AddCommand(validatorStatusCmd)
	rootCmd.AddCommand(validatorCmd)
}
