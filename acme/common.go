package acme

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
)

func GetPrivateKey(path string) (crypto.PrivateKey, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, fmt.Errorf("failed to generate RSA key: %w", err)
		}

		pem := certcrypto.PEMEncode(key)
		if err := os.WriteFile(path, pem, 0600); err != nil {
			return nil, fmt.Errorf("failed to write %s: %w", path, err)
		}

		return key, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	return certcrypto.ParsePEMPrivateKey(data)
}
