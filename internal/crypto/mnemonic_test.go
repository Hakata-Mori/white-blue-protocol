package crypto

import (
	"strings"
	"testing"
)

func TestGenerateKeyPairFromMnemonic(t *testing.T) {
	mnemonic, kp, err := GenerateKeyPairFromMnemonic()
	if err != nil {
		t.Fatalf("GenerateKeyPairFromMnemonic failed: %v", err)
	}

	words := strings.Fields(mnemonic)
	if len(words) != 12 {
		t.Fatalf("expected 12 words, got %d", len(words))
	}

	if kp.PrivateKey == "" || kp.PublicKey == "" || kp.Address == "" {
		t.Fatal("keypair fields must not be empty")
	}

	if !strings.HasPrefix(kp.Address, "0x") || len(kp.Address) != 42 {
		t.Fatalf("invalid address format: %s", kp.Address)
	}
}

func TestRecoverKeyPairFromMnemonic(t *testing.T) {
	mnemonic, kp1, err := GenerateKeyPairFromMnemonic()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	kp2, err := RecoverKeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("recover failed: %v", err)
	}

	if kp1.PrivateKey != kp2.PrivateKey {
		t.Fatal("private keys do not match")
	}
	if kp1.PublicKey != kp2.PublicKey {
		t.Fatal("public keys do not match")
	}
	if kp1.Address != kp2.Address {
		t.Fatal("addresses do not match")
	}
}

func TestRecoverDeterministic(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	kp1, err := RecoverKeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("first recover failed: %v", err)
	}

	kp2, err := RecoverKeyPairFromMnemonic(mnemonic)
	if err != nil {
		t.Fatalf("second recover failed: %v", err)
	}

	if kp1.Address != kp2.Address {
		t.Fatal("same mnemonic produced different addresses")
	}
}

func TestRecoverInvalidMnemonic(t *testing.T) {
	_, err := RecoverKeyPairFromMnemonic("invalid words that are not a mnemonic")
	if err == nil {
		t.Fatal("expected error for invalid mnemonic")
	}
}

func TestRecoverEmptyMnemonic(t *testing.T) {
	_, err := RecoverKeyPairFromMnemonic("")
	if err == nil {
		t.Fatal("expected error for empty mnemonic")
	}
}

func TestMnemonicKeyCanSign(t *testing.T) {
	_, kp, err := GenerateKeyPairFromMnemonic()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	msg := []byte("test message")
	sig, err := Sign(kp.PrivateKey, msg)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	valid, err := Verify(kp.PublicKey, msg, sig)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !valid {
		t.Fatal("signature verification failed")
	}
}
