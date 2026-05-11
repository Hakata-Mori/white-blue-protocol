package state

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/white-blue-protocol/wblue/internal/types"
)

func TestVerifyJoinPoWValid(t *testing.T) {
	types.SetDevMode()
	address := "0xtestaddr"
	height := uint64(100)

	var nonce uint64
	requiredBytes := types.GetPoWDifficulty() / 8
	for nonce = 0; nonce < 1_000_000; nonce++ {
		data := fmt.Sprintf("%s:%d:%d", address, height, nonce)
		hash := sha256.Sum256([]byte(data))
		valid := true
		for i := 0; i < requiredBytes; i++ {
			if hash[i] != 0 {
				valid = false
				break
			}
		}
		if valid {
			break
		}
	}

	payload, _ := json.Marshal(struct {
		Nonce uint64 `json:"nonce"`
	}{Nonce: nonce})

	if !verifyJoinPoW(address, height, payload) {
		t.Fatal("valid PoW should pass verification")
	}
}

func TestVerifyJoinPoWInvalid(t *testing.T) {
	types.SetDevMode()

	payload, _ := json.Marshal(struct {
		Nonce uint64 `json:"nonce"`
	}{Nonce: 99999999})

	if verifyJoinPoW("0xtest", 1, payload) {
		t.Fatal("invalid PoW nonce should fail")
	}
}

func TestVerifyJoinPoWEmptyPayload(t *testing.T) {
	if verifyJoinPoW("0xtest", 1, nil) {
		t.Fatal("empty payload should fail")
	}
	if verifyJoinPoW("0xtest", 1, []byte{}) {
		t.Fatal("empty payload should fail")
	}
}

func TestVerifyJoinPoWBadJSON(t *testing.T) {
	if verifyJoinPoW("0xtest", 1, []byte("not json")) {
		t.Fatal("bad json should fail")
	}
}
