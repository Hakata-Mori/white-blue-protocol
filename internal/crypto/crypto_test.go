package crypto

import (
	"encoding/hex"
	"testing"
)

func TestGenerateKeyPairRoundtrip(t *testing.T) {
	kp, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	if kp.PrivateKey == "" || kp.PublicKey == "" || kp.Address == "" {
		t.Fatal("keypair fields should not be empty")
	}
	if len(kp.Address) != 42 {
		t.Fatalf("address should be 42 chars (0x + 40 hex), got %d", len(kp.Address))
	}
	if kp.Address[:2] != "0x" {
		t.Fatal("address should start with 0x")
	}

	privBytes, _ := hex.DecodeString(kp.PrivateKey)
	if len(privBytes) != 32 {
		t.Fatalf("private key should be 32 bytes, got %d", len(privBytes))
	}
}

func TestSignVerifyRoundtrip(t *testing.T) {
	kp, _ := GenerateKeyPair()
	data := []byte("hello blockchain")

	sig, err := Sign(kp.PrivateKey, data)
	if err != nil {
		t.Fatalf("sign failed: %v", err)
	}

	ok, err := Verify(kp.PublicKey, data, sig)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !ok {
		t.Fatal("valid signature should verify")
	}
}

func TestVerifyWrongData(t *testing.T) {
	kp, _ := GenerateKeyPair()
	sig, _ := Sign(kp.PrivateKey, []byte("original"))

	ok, err := Verify(kp.PublicKey, []byte("tampered"), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("signature should not verify with different data")
	}
}

func TestVerifyWrongKey(t *testing.T) {
	kp1, _ := GenerateKeyPair()
	kp2, _ := GenerateKeyPair()
	data := []byte("test")

	sig, _ := Sign(kp1.PrivateKey, data)

	ok, _ := Verify(kp2.PublicKey, data, sig)
	if ok {
		t.Fatal("signature should not verify with wrong public key")
	}
}

func TestVerifyBadSignatureLength(t *testing.T) {
	kp, _ := GenerateKeyPair()
	_, err := Verify(kp.PublicKey, []byte("test"), "aabb")
	if err == nil {
		t.Fatal("short signature should fail")
	}
}

func TestVerifyBadSignatureHex(t *testing.T) {
	kp, _ := GenerateKeyPair()
	_, err := Verify(kp.PublicKey, []byte("test"), "not_hex!!")
	if err == nil {
		t.Fatal("invalid hex should fail")
	}
}

func TestSignBadPrivateKey(t *testing.T) {
	_, err := Sign("zzzz", []byte("test"))
	if err == nil {
		t.Fatal("invalid hex private key should fail")
	}
}

func TestVerifyBadPublicKey(t *testing.T) {
	_, err := Verify("zzzz", []byte("test"), "00")
	if err == nil {
		t.Fatal("invalid hex public key should fail")
	}
}

func TestPublicKeyFromHexInvalid(t *testing.T) {
	_, err := PublicKeyFromHex("0000000000000000000000000000000000")
	if err == nil {
		t.Fatal("invalid compressed public key should fail")
	}
}

func TestVerifyWithAddressValid(t *testing.T) {
	kp, _ := GenerateKeyPair()
	data := []byte("test verify with address")
	sig, _ := Sign(kp.PrivateKey, data)

	ok, err := VerifyWithAddress(kp.PublicKey, kp.Address, data, sig)
	if err != nil {
		t.Fatalf("should succeed: %v", err)
	}
	if !ok {
		t.Fatal("valid signature+address should verify")
	}
}

func TestVerifyWithAddressMismatch(t *testing.T) {
	kp, _ := GenerateKeyPair()
	data := []byte("test")
	sig, _ := Sign(kp.PrivateKey, data)

	_, err := VerifyWithAddress(kp.PublicKey, "0xwrongaddress000000000000000000000000dead", data, sig)
	if err == nil {
		t.Fatal("mismatched address should fail")
	}
}

func TestPubKeyToAddressDeterministic(t *testing.T) {
	kp, _ := GenerateKeyPair()
	pubBytes, _ := hex.DecodeString(kp.PublicKey)

	addr1 := PubKeyToAddress(pubBytes)
	addr2 := PubKeyToAddress(pubBytes)
	if addr1 != addr2 {
		t.Fatal("address derivation should be deterministic")
	}
	if addr1 != kp.Address {
		t.Fatal("derived address should match keypair address")
	}
}

func TestPubKeyToAddressFormat(t *testing.T) {
	addr := PubKeyToAddress([]byte("some_public_key_bytes"))
	if len(addr) != 42 {
		t.Fatalf("address should be 42 chars, got %d", len(addr))
	}
	if addr[:2] != "0x" {
		t.Fatal("address should start with 0x")
	}
}

func TestSHA256HexDeterministic(t *testing.T) {
	h1 := SHA256Hex([]byte("hello"))
	h2 := SHA256Hex([]byte("hello"))
	if h1 != h2 {
		t.Fatal("same input should produce same hash")
	}
	if h1 == SHA256Hex([]byte("world")) {
		t.Fatal("different inputs should produce different hashes")
	}
	if len(h1) != 64 {
		t.Fatalf("hex hash should be 64 chars, got %d", len(h1))
	}
}

func TestMerkleRootEmpty(t *testing.T) {
	root := MerkleRoot(nil)
	if root == "" {
		t.Fatal("empty merkle root should not be empty string")
	}
	root2 := MerkleRoot([]string{})
	if root != root2 {
		t.Fatal("nil and empty should produce same root")
	}
}

func TestMerkleRootSingle(t *testing.T) {
	root := MerkleRoot([]string{"abc123"})
	if root == "" {
		t.Fatal("single element merkle root should not be empty")
	}
}

func TestMerkleRootOddElements(t *testing.T) {
	root3 := MerkleRoot([]string{"a", "b", "c"})
	if root3 == "" {
		t.Fatal("odd element merkle root should not be empty")
	}
	root2 := MerkleRoot([]string{"a", "b"})
	if root3 == root2 {
		t.Fatal("3 elements and 2 elements should produce different root")
	}
}

func TestMerkleRootDeterministic(t *testing.T) {
	hashes := []string{"h1", "h2", "h3", "h4"}
	r1 := MerkleRoot(hashes)
	r2 := MerkleRoot(hashes)
	if r1 != r2 {
		t.Fatal("same input should produce same merkle root")
	}
}

func TestMerkleRootOrderMatters(t *testing.T) {
	r1 := MerkleRoot([]string{"a", "b"})
	r2 := MerkleRoot([]string{"b", "a"})
	if r1 == r2 {
		t.Fatal("different order should produce different root")
	}
}

func TestLowSNormalization(t *testing.T) {
	kp, _ := GenerateKeyPair()
	data := []byte("test low-s normalization")

	for i := 0; i < 20; i++ {
		sig, err := Sign(kp.PrivateKey, data)
		if err != nil {
			t.Fatal(err)
		}
		ok, err := Verify(kp.PublicKey, data, sig)
		if err != nil {
			t.Fatalf("verify failed on iteration %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("signature should verify on iteration %d", i)
		}
	}
}
