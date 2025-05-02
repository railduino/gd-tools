package main

import (
	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandEnable, "devOnly")
}

var commandEnable = &cli.Command{
	Name:        "enable",
	Usage:       T("enable-cmd-usage"),
	Description: T("enable-cmd-describe"),
	Flags:       []cli.Flag{},
	Action:      runEnable,
}

func runEnable(c *cli.Context) error {
	// TODO set enable flag in labels

	return nil
}
