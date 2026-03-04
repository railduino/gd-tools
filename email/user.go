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

	// Forwarding (dev-only config, rendered into sieve_after)
	ForwardKeepCopy *bool    `json:"forward_keep_copy,omitempty"` // true => redirect :copy
	Forwards        []string `json:"forwards"`                    // 0..n forwarding targets
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

func (u User) HasForwards() bool {
	return len(u.Forwards) > 0
}

func (u User) ForwardKeep() bool {
	if len(u.Forwards) == 0 {
		return false
	}

	if u.ForwardKeepCopy == nil {
		return true // default: keep local copy
	}

	return *u.ForwardKeepCopy
}

func (u *User) NormalizeAndValidateForwards() error {
	seen := map[string]bool{}
	out := make([]string, 0, len(u.Forwards))

	for _, f := range u.Forwards {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		pa, err := mail.ParseAddress(f)
		if err != nil {
			return fmt.Errorf("invalid forward address '%s' for %s: %w", f, u.Email(), err)
		}
		addr := strings.ToLower(pa.Address)
		if addr == strings.ToLower(u.Email()) {
			return fmt.Errorf("forward loop: %s forwards to itself", u.Email())
		}
		if !seen[addr] {
			seen[addr] = true
			out = append(out, addr)
		}
	}

	u.Forwards = out
	return nil
}
