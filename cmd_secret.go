package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandSecret, "any")
}

var commandSecret = &cli.Command{
	Name:        "secret",
	Usage:       T("secret-cmd-usage"),
	Description: T("secret-cmd-describe"),
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "size",
			Usage: "Byte-Länge des Schlüssels vor Base64-Kodierung",
			Value: 32,
		},
	},
	Action: runSecret,
}

func runSecret(c *cli.Context) error {
	size := c.Int("size")
	key := make([]byte, size)

	if _, err := rand.Read(key); err != nil {
		return fmt.Errorf("Fehler beim Erzeugen des Schlüssels: %w", err)
	}

	fmt.Println(base64.StdEncoding.EncodeToString(key))
	return nil
}
