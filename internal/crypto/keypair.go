package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/white-blue-protocol/wblue/internal/types"
)

func GenerateKeyPair() (*types.KeyPair, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	privBytes := privateKey.D.Bytes()
	pubBytes := elliptic.MarshalCompressed(privateKey.PublicKey.Curve, privateKey.PublicKey.X, privateKey.PublicKey.Y)

	address := PubKeyToAddress(pubBytes)

	return &types.KeyPair{
		PrivateKey: hex.EncodeToString(privBytes),
		PublicKey:  hex.EncodeToString(pubBytes),
		Address:    address,
	}, nil
}

func PrivateKeyFromHex(hexStr string) (*ecdsa.PrivateKey, error) {
	privBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	curve := elliptic.P256()
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = curve
	priv.D = new(big.Int).SetBytes(privBytes)
	priv.PublicKey.X, priv.PublicKey.Y = curve.ScalarBaseMult(privBytes)

	return priv, nil
}

func PublicKeyFromHex(hexStr string) (*ecdsa.PublicKey, error) {
	pubBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}

	curve := elliptic.P256()
	x, y := elliptic.UnmarshalCompressed(curve, pubBytes)
	if x == nil {
		return nil, fmt.Errorf("invalid public key")
	}

	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}
