package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func SHA256Hex(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func MerkleRoot(txHashes []string) string {
	if len(txHashes) == 0 {
		return SHA256Hex(nil)
	}

	hashes := make([]string, len(txHashes))
	copy(hashes, txHashes)

	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var nextLevel []string
		for i := 0; i < len(hashes); i += 2 {
			combined := fmt.Sprintf("%s%s", hashes[i], hashes[i+1])
			nextLevel = append(nextLevel, SHA256Hex([]byte(combined)))
		}
		hashes = nextLevel
	}

	return hashes[0]
}
