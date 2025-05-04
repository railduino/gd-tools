package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

// Registered project kinds (eg. traefik, wordpress, nextcloud, mail_host)
var generateKinds []*cli.Command

var generateCommand *cli.Command = commandGenerate

func init() {
	AddSubCommand(commandGenerate, "dev")
}

var generateDependsFlag = cli.StringSliceFlag{
	Name:    "depends",
	Aliases: []string{"d"},
	Usage:   T("generate-flag-depends"),
}

func RegisterProjectKind(cmd *cli.Command) {
	generateKinds = append(generateKinds, cmd)
	generateCommand.Subcommands = append(generateCommand.Subcommands, cmd)

	sort.Slice(generateKinds, func(i, j int) bool {
		return generateKinds[i].Name < generateKinds[j].Name
	})
}

var commandGenerate = &cli.Command{
	Name:        "generate",
	Usage:       T("generate-cmd-usage"),
	Description: T("generate-cmd-describe"),
	Aliases:     []string{"g"},
	Subcommands: []*cli.Command{},
	Action:      runGenerateDispatcher,
}

func runGenerateDispatcher(c *cli.Context) error {
	fmt.Println(T("generate-cmd-list-kinds"))

	for _, sub := range generateKinds {
		if sub.Name == "help" {
			continue
		}
		fmt.Printf("  %-20s %s\n", sub.Name, sub.Usage)
	}

	return nil
}

func GenerateEnvFromUIDs(envPath string) error {
	data, err := os.ReadFile(filepath.Join(LetsEncryptDir, "uids.json"))
	if err != nil {
		return err
	}

	var uids map[string]string
	if err := json.Unmarshal(data, &uids); err != nil {
		return fmt.Errorf("uids.json ungültig: %w", err)
	}

	lines := []string{
		"GDTOOLS_UID=" + uids["gd-tools.uid"],
		"GDTOOLS_GID=" + uids["gd-tools.gid"],
		"DOCKER_GID=" + uids["docker.gid"],
	}

	return os.WriteFile(envPath, []byte(strings.Join(lines, "\n")+"\n"), 0644)
}
