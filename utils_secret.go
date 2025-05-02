package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Secret struct {
	Domain string `json:"domain"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

// GetSecret returns a secret value for a given key
func GetSecret(key, mode string) (string, error) {
	switch mode {
	case "", "bcrypt":
		return generateBcrypt(key)
	default:
		msg := Tf("secret-err-unknown-mode", mode)
		return "", fmt.Errorf(msg)
	}
}

func generateBcrypt(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf(T("secret-err-empty"))
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}
