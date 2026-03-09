package utils

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hash"
	"math/big"
	"os"
	"os/exec"
	"sort"

	"github.com/tv42/zbase32"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/pbkdf2"
)

const (
	SecretsName  = "secrets.json"
	MailUserName = "mailuser"
)

type Secret struct {
	Domain string `json:"domain"`
	User   string `json:"user"`
	Input  string `json:"input"`
	Output string `json:"output"`
}

type SecretList struct {
	Secrets []Secret `json:"entries"`
}

func GenerateSecret(key, mode string) (string, error) {
	switch mode {
	case "bcrypt", "":
		return GenerateBcrypt(key)
	case "pbkdf2":
		return GeneratePBKDF2(key)
	}

	return "", fmt.Errorf("unknown secret mode %s", mode)
}

func GenerateBcrypt(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("empty passwords are not allowed")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func GeneratePBKDF2(password string) (string, error) {
	salt := make([]byte, 20)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	h := func() hash.Hash { x, _ := blake2b.New512(nil); return x }
	key := pbkdf2.Key([]byte(password), salt, 16000, 32, h)

	return fmt.Sprintf("$1$%s$%s",
		zbase32.EncodeToString(salt),
		zbase32.EncodeToString(key),
	), nil
}

func LoadSecrets() (*SecretList, error) {
	var list SecretList

	content, err := os.ReadFile(SecretsName)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		} else {
			return nil, err
		}
	}

	if err := json.Unmarshal(content, &list); err != nil {
		return nil, err
	}

	return &list, nil
}

func (list *SecretList) Save() error {
	sort.Slice(list.Secrets, func(i, j int) bool {
		if list.Secrets[i].Domain == list.Secrets[j].Domain {
			return list.Secrets[i].User < list.Secrets[j].User
		}
		return list.Secrets[i].Domain < list.Secrets[j].Domain
	})

	content, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", SecretsName, err)
	}

	existing, err := os.ReadFile(SecretsName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(SecretsName, content, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", SecretsName, err)
	}

	return nil
}

func (list *SecretList) Get(domain, user string) *Secret {
	for _, entry := range list.Secrets {
		if entry.Domain == domain && entry.User == user {
			return &entry
		}
	}

	return nil
}

func (list *SecretList) SetMailUser(address, password string) (string, string, error) {
	if password == "" {
		password, _ = CreatePassword(20)
	}

	output, err := GenerateSecret(password, "bcrypt")
	if err != nil {
		return "", "", err
	}

	if err := list.Set(MailUserName, address, password, output); err != nil {
		return "", "", err
	}

	return password, output, nil
}

func (list *SecretList) Set(domain, user, input, output string) error {
	if entry := list.Get(domain, user); entry != nil {
		if entry.Input != input {
			entry.Input = input
			entry.Output = output
		}
	} else {
		entry := Secret{
			Domain: domain,
			User:   user,
			Input:  input,
			Output: output,
		}
		list.Secrets = append(list.Secrets, entry)
	}

	return list.Save()
}

func CreatePassword(length int) (string, error) {
	charset := "abcdefghijkmnopqrstuvwxyzABCDEFGHIJKLMNPQRSTUVWXYZ0123456789"

	pass := make([]byte, length)
	for i := range pass {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		pass[i] = charset[n.Int64()]
	}
	return string(pass), nil
}

func FetchPassword(length int, domain, user string) (string, error) {
	list, err := LoadSecrets()
	if err != nil {
		return "", err
	}

	if entry := list.Get(domain, user); entry != nil {
		return entry.Output, nil
	}

	password, err := CreatePassword(length)
	if err != nil {
		return "", err
	}

	if err := list.Set(domain, user, "", password); err != nil {
		return "", err
	}

	return password, nil
}

func GetRSAKeyPair(fqdn string) ([]byte, []byte, error) {
	priv, err1 := os.ReadFile("root_id_rsa")
	publ, err2 := os.ReadFile("root_id_rsa.pub")
	if err1 == nil && err2 == nil {
		return priv, publ, nil
	}

	cmd := exec.Command(
		"ssh-keygen",
		"-t", "rsa",
		"-C", "root@"+fqdn,
		"-f", "root_id_rsa",
		"-N", "",
	)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, nil, err
	}

	priv, err1 = os.ReadFile("root_id_rsa")
	if err1 != nil {
		return nil, nil, err1
	}
	publ, err2 = os.ReadFile("root_id_rsa.pub")
	if err2 != nil {
		return nil, nil, err2
	}

	return priv, publ, nil
}
