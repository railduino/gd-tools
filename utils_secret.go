package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
)

type Secret struct {
	Domain string `json:"domain"`
	User   string `json:"user"`
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

func LoadSecret(projectPath string, domain string, user string) (*Secret, error) {
	secretsPath := filepath.Join(projectPath, "secrets.json")
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return nil, err
	}

	var secrets []Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}

	for _, s := range secrets {
		if s.Domain == domain && s.User == user {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("kein Secret für Domain '%s' und User '%s' gefunden", domain, user)
}

func SaveSecrets(projectPath string, secrets []Secret) error {
	secretsPath := filepath.Join(projectPath, "secrets.json")

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(secretsPath, data, 0600)
}

func GenerateRandomPassword() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!$%&*+-_"
	const length = 16

	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // darf in dev abbrechen
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}
