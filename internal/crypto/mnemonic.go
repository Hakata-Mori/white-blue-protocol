package crypto

import (
	"crypto/elliptic"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/tyler-smith/go-bip39"
	"github.com/white-blue-protocol/wblue/internal/types"
)

func GenerateKeyPairFromMnemonic() (string, *types.KeyPair, error) {
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", nil, fmt.Errorf("generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", nil, fmt.Errorf("generate mnemonic: %w", err)
	}

	kp, err := RecoverKeyPairFromMnemonic(mnemonic)
	if err != nil {
		return "", nil, err
	}

	return mnemonic, kp, nil
}

func RecoverKeyPairFromMnemonic(mnemonic string) (*types.KeyPair, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	seed := bip39.NewSeed(mnemonic, "")

	curve := elliptic.P256()
	n := curve.Params().N

	privInt := new(big.Int).SetBytes(seed[:32])
	privInt.Mod(privInt, n)

	if privInt.Sign() == 0 {
		return nil, fmt.Errorf("derived private key is zero (try different mnemonic)")
	}

	privBytes := privInt.Bytes()
	if len(privBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(privBytes):], privBytes)
		privBytes = padded
	}

	x, y := curve.ScalarBaseMult(privBytes)
	pubBytes := elliptic.MarshalCompressed(curve, x, y)
	address := PubKeyToAddress(pubBytes)

	return &types.KeyPair{
		PrivateKey: hex.EncodeToString(privBytes),
		PublicKey:  hex.EncodeToString(pubBytes),
		Address:    address,
	}, nil
}
