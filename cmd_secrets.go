package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandSecrets, "dev")
}

var commandSecrets = &cli.Command{
	Name:        "secrets",
	Usage:       T("secrets-cmd-usage"),
	Description: T("secrets-cmd-describe"),
	ArgsUsage:   "[projekt]",
	Action:      runSecretsList,
}

func runSecretsList(c *cli.Context) error {
	showPlain := CheckEnv("dev")

	// Einzelprojekt-Modus
	if c.NArg() >= 1 {
		projectName := c.Args().First()
		return showSecretsForProject(projectName, showPlain)
	}

	// Multi-Projekt-Modus
	err := filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return nil // ignorieren
		}
		if info.Name() == "secrets.json" && info.Mode().IsRegular() {
			project := filepath.Dir(path)
			return showSecretsWithPrefix(project, path, showPlain)
		}
		return nil
	})
	return err
}

func showSecretsWithPrefix(project, secretsPath string, showPlain bool) error {
	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return nil // still, bei Fehler einfach überspringen
	}

	var secrets []Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil
	}

	fmt.Printf("%-15s | %-25s | %-25s | %-15s | %-60s\n", "PROJEKT", "DOMAIN", "USER", "PLAINTEXT", "HASH")
	fmt.Println(strings.Repeat("-", 140))
	for _, s := range secrets {
		plain := ""
		if showPlain {
			plain = s.Input
		}
		fmt.Printf("%-15s | %-25s | %-25s | %-16s | %-60s\n", project, s.Domain, s.User, plain, s.Output)
	}

	return nil
}

func showSecretsForProject(projectName string, showPlain bool) error {
	secretsPath := filepath.Join(projectName, "secrets.json")

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return fmt.Errorf(Tf("secrets-err-read-failed", secretsPath, err))
	}

	var secrets []Secret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return fmt.Errorf(Tf("secrets-err-parse-failed", secretsPath, err))
	}

	fmt.Printf("%-25s | %-25s | %-16s | %-60s\n", "DOMAIN", "USER", "PLAINTEXT", "HASH")
	fmt.Println(strings.Repeat("-", 120))
	for _, s := range secrets {
		plain := ""
		if showPlain {
			plain = s.Input
		}
		fmt.Printf("%-25s | %-25s | %-15s | %-60s\n", s.Domain, s.User, plain, s.Output)
	}
	return nil
}
