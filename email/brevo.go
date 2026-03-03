package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	BrevoName = "brevo.json"
)

type Brevo struct {
	Enabled bool   `json:"enabled"` // Active (or not)
	Server  string `json:"server"`  // SMTP Server
	Port    int    `json:"port"`    // SMTP Port
	ID      string `json:"id"`      // (Email) Identifier
	Code    string `json:"code"`    // (Domain) Verification Code
	Key     string `json:"key"`     // SASL Password (postmap)
	DMARC   string `json:"dmarc"`   // DMARC Redirect
}

func ReadBrevo(c *cli.Context) (*Brevo, error) {
	content, err := os.ReadFile(BrevoName)
	if err != nil {
		if os.IsNotExist(err) {
			brevo := Brevo{
				Enabled: c.Bool("enabled"),
				Server:  c.String("server"),
				Port:    c.Int("port"),
				ID:      c.String("id"),
				Code:    c.String("code"),
				Key:     c.String("key"),
				DMARC:   c.String("dmarc"),
			}
			return &brevo, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", BrevoName, err)
	}

	var brevo Brevo
	if err := json.Unmarshal(content, &brevo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", BrevoName, err)
	}

	return &brevo, nil
}

func (brevo *Brevo) Save() error {
	content, err := json.MarshalIndent(brevo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", BrevoName, err)
	}

	existing, err := os.ReadFile(BrevoName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(BrevoName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", BrevoName, err)
	}

	return nil
}
