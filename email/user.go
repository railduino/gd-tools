package email

import (
	"fmt"
	"net/mail"
	"strings"
)

type User struct {
	Local    string   `json:"local"`    // Local part of the email address
	Domain   string   `json:"domain"`   // Domain part of the email address
	Name     string   `json:"name"`     // Display name of the user
	Password string   `json:"password"` // Plaintext or hashed password
	Locked   bool     `json:"locked"`   // Whether the account is locked
	Aliases  []string `json:"aliases"`  // Additional aliases for this user
	Quota    string   `json:"quota"`    // Optional mailbox quota

	Forwards []string `json:"forwards,omitempty"` // 0..n forwarding targets
	Dismiss  bool     `json:"dismiss,omitempty"`  // if true: do not keep local copy
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

func (u *User) AddForward(addr string) error {
	parsed, err := mail.ParseAddress(strings.TrimSpace(addr))
	if err != nil {
		return fmt.Errorf("invalid forward address '%s': %w", addr, err)
	}

	target := strings.ToLower(parsed.Address)
	self := strings.ToLower(u.Email())
	if target == self {
		return fmt.Errorf("forward to itself is not allowed")
	}

	for _, f := range u.Forwards {
		if strings.ToLower(strings.TrimSpace(f)) == target {
			return nil // ignore duplicates
		}
	}
	u.Forwards = append(u.Forwards, target)

	return nil
}

func (u *User) DeleteForward(addr string) error {
	parsed, err := mail.ParseAddress(strings.TrimSpace(addr))
	if err != nil {
		return fmt.Errorf("invalid forward address '%s': %w", addr, err)
	}
	target := strings.ToLower(parsed.Address)

	out := u.Forwards[:0]
	found := false
	for _, f := range u.Forwards {
		if strings.ToLower(strings.TrimSpace(f)) == target {
			found = true
			continue
		}
		out = append(out, f)
	}
	u.Forwards = out

	if !found {
		return fmt.Errorf("forward target not found: %s", target)
	}

	return nil
}

func (u *User) ClearForwards() {
	u.Forwards = nil
	u.Dismiss = false
}
