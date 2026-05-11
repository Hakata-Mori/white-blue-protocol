package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"
)

type KDFParams struct {
	N    int    `json:"n"`
	R    int    `json:"r"`
	P    int    `json:"p"`
	DKLen int   `json:"dklen"`
	Salt string `json:"salt"`
}

type CryptoData struct {
	Cipher     string    `json:"cipher"`
	CipherText string    `json:"ciphertext"`
	Nonce      string    `json:"nonce"`
	KDF        string    `json:"kdf"`
	KDFParams  KDFParams `json:"kdfparams"`
}

type KeystoreFile struct {
	Version   int        `json:"version"`
	Address   string     `json:"address"`
	PublicKey string     `json:"publicKey"`
	Crypto    CryptoData `json:"crypto"`
}

func Encrypt(privKeyHex, publicKeyHex, address, password string) (*KeystoreFile, error) {
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	dk, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	privBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, privBytes, nil)

	return &KeystoreFile{
		Version:   1,
		Address:   address,
		PublicKey: publicKeyHex,
		Crypto: CryptoData{
			Cipher:     "aes-256-gcm",
			CipherText: hex.EncodeToString(ciphertext),
			Nonce:      hex.EncodeToString(nonce),
			KDF:        "scrypt",
			KDFParams: KDFParams{
				N:    32768,
				R:    8,
				P:    1,
				DKLen: 32,
				Salt: hex.EncodeToString(salt),
			},
		},
	}, nil
}

func Decrypt(ks *KeystoreFile, password string) (string, error) {
	salt, err := hex.DecodeString(ks.Crypto.KDFParams.Salt)
	if err != nil {
		return "", err
	}

	dk, err := scrypt.Key([]byte(password), salt,
		ks.Crypto.KDFParams.N,
		ks.Crypto.KDFParams.R,
		ks.Crypto.KDFParams.P,
		ks.Crypto.KDFParams.DKLen,
	)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(dk)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce, err := hex.DecodeString(ks.Crypto.Nonce)
	if err != nil {
		return "", err
	}

	ciphertext, err := hex.DecodeString(ks.Crypto.CipherText)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("wrong password or corrupted keystore")
	}

	return hex.EncodeToString(plaintext), nil
}

func Save(ks *KeystoreFile, path string) error {
	data, err := json.MarshalIndent(ks, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func Load(path string) (*KeystoreFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var ks KeystoreFile
	if err := json.Unmarshal(data, &ks); err != nil {
		return nil, err
	}
	return &ks, nil
}

func IsKeystoreFile(data []byte) bool {
	var ks KeystoreFile
	if err := json.Unmarshal(data, &ks); err != nil {
		return false
	}
	return ks.Version > 0 && ks.Crypto.Cipher != ""
}
