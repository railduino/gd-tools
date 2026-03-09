package account

import (
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/domain"
	"github.com/urfave/cli/v2"
)

var Describe = `The account command handles Email accounts.`

var Command = &cli.Command{
	Name:        "account",
	Usage:       "Handle email users (see 'domain' for email domains)",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Subcommands: []*cli.Command{
		AddCommand,
		BrevoCommand,
		DeployCommand,
		ForwardCommand,
		domain.ListCommand,
	},
	Action: domain.ListRun,
}
