package main

import (
	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandHash, "any")
}

var commandHash = &cli.Command{
	Name:        "hash",
	Usage:       T("hash-cmd-usage"),
	Description: T("hash-cmd-describe"),
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "crc32",
			Aliases: []string{"c"},
			Usage:   T("usage-command-hash-crc32"),
		},
	},
	Action: runHash,
}

func runHash(c *cli.Context) error {
	return nil
}
