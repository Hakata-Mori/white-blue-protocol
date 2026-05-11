package keystore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptRoundtrip(t *testing.T) {
	privKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	pubKey := "04abcdef"
	addr := "0xtest"
	password := "mysecretpassword"

	ks, err := Encrypt(privKey, pubKey, addr, password)
	if err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	if ks.Version != 1 {
		t.Fatalf("expected version 1, got %d", ks.Version)
	}
	if ks.Address != addr {
		t.Fatalf("address mismatch")
	}
	if ks.Crypto.Cipher != "aes-256-gcm" {
		t.Fatalf("cipher mismatch")
	}

	decrypted, err := Decrypt(ks, password)
	if err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}
	if decrypted != privKey {
		t.Fatalf("decrypted key mismatch: expected %s, got %s", privKey, decrypted)
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	privKey := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

	ks, err := Encrypt(privKey, "04pub", "0xaddr", "correct")
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(ks, "wrong")
	if err == nil {
		t.Fatal("decryption with wrong password should fail")
	}
}

func TestSaveAndLoad(t *testing.T) {
	privKey := "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789"

	ks, err := Encrypt(privKey, "04pub", "0xaddr", "pw123")
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "wallet.json")

	if err := Save(ks, path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	decrypted, err := Decrypt(loaded, "pw123")
	if err != nil {
		t.Fatalf("decrypt loaded failed: %v", err)
	}
	if decrypted != privKey {
		t.Fatal("loaded key mismatch")
	}
}

func TestIsKeystoreFile(t *testing.T) {
	ks, _ := Encrypt("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "04pub", "0x", "pw")

	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	Save(ks, path)

	data, _ := os.ReadFile(path)
	if !IsKeystoreFile(data) {
		t.Fatal("should detect valid keystore file")
	}

	if IsKeystoreFile([]byte(`{"foo":"bar"}`)) {
		t.Fatal("should reject non-keystore JSON")
	}

	if IsKeystoreFile([]byte("not json")) {
		t.Fatal("should reject non-JSON data")
	}
}
