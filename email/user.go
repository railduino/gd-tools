package email

import (
	"fmt"
	"net/mail"
	"strings"
)

const (
	SecretDomain = "mailuser"
)

type User struct {
	Local    string   `json:"local"`    // Local part of the email address
	Domain   string   `json:"domain"`   // Domain part of the email address
	Name     string   `json:"name"`     // Display name of the user
	Password string   `json:"password"` // Plaintext or hashed password
	Locked   bool     `json:"locked"`   // Whether the account is locked
	Aliases  []string `json:"aliases"`  // Additional aliases for this user
	Quota    string   `json:"quota"`    // Optional mailbox quota
}

func (u User) Email() string {
	return u.Local + "@" + u.Domain
}

func (u *User) Address() string {
	if u.Name == "" {
		return u.Email()
	}
	return fmt.Sprintf("%s <%s@%s>", u.Name, u.Local, u.Domain)
}

func MakeUser(addr string) (*User, error) {
	parsed, err := mail.ParseAddress(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to validate address '%s': %w", addr, err)
	}

	parts := strings.Split(parsed.Address, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("impossible email address '%s'", parsed.Address)
	}

	user := User{
		Name:   parsed.Name,
		Local:  parts[0],
		Domain: parts[1],
	}

	return &user, nil
}
