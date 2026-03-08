package account

import (
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The account command handles Email accounts.`

var Command = &cli.Command{
	Name:        "account",
	Usage:       "Handle email domains and user accounts",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Subcommands: []*cli.Command{
		AddCommand,
		AliasCommand,
		BrevoCommand,
		CAACommand,
		DeployCommand,
		ForwardCommand,
		ListCommand,
		SPFCommand,
		SyncCommand,
	},
	Action: ListRun,
}
