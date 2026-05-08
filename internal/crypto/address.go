package crypto

import (
	"crypto/sha256"
	"encoding/hex"
)

func PubKeyToAddress(pubKeyBytes []byte) string {
	hash := sha256.Sum256(pubKeyBytes)
	return "0x" + hex.EncodeToString(hash[12:])
}
