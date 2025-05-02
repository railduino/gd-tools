package main

import (
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"
)

var generateDependsFlag = cli.StringSliceFlag{
	Name:    "depends",
	Aliases: []string{"d"},
	Usage:   T("project-flag-depends"),
}

// Registered project kinds (eg. traefik, wordpress, nextcloud, mail_host)
var generateKinds []*cli.Command

var generateCommand *cli.Command = commandGenerate

func init() {
	AddSubCommand(commandGenerate, "devOnly")
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
