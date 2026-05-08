package crypto

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

func Sign(privateKeyHex string, data []byte) (string, error) {
	privKey, err := PrivateKeyFromHex(privateKeyHex)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(rand.Reader, privKey, hash[:])
	if err != nil {
		return "", err
	}

	rBytes := r.Bytes()
	sBytes := s.Bytes()
	sig := make([]byte, 64)
	copy(sig[32-len(rBytes):32], rBytes)
	copy(sig[64-len(sBytes):64], sBytes)

	return hex.EncodeToString(sig), nil
}

func Verify(publicKeyHex string, data []byte, signatureHex string) (bool, error) {
	pubKey, err := PublicKeyFromHex(publicKeyHex)
	if err != nil {
		return false, err
	}

	sigBytes, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false, err
	}

	if len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature length")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	hash := sha256.Sum256(data)
	return ecdsa.Verify(pubKey, hash[:], r, s), nil
}

func VerifyWithAddress(publicKeyHex string, address string, data []byte, signatureHex string) (bool, error) {
	pubBytes, err := hex.DecodeString(publicKeyHex)
	if err != nil {
		return false, err
	}

	derivedAddr := PubKeyToAddress(pubBytes)
	if derivedAddr != address {
		return false, fmt.Errorf("public key does not match address")
	}

	return Verify(publicKeyHex, data, signatureHex)
}
