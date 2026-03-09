package domain

import (
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The domain command handles email domains.`

var Command = &cli.Command{
	Name:        "domain",
	Usage:       "Handle email domains (see 'account' for email users)",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Subcommands: []*cli.Command{
		AliasCommand,
		// BrevoCommand,
		CAACommand,
		// DeployCommand,
		ListCommand,
		SPFCommand,
		SyncCommand,
	},
	Action: ListRun,
}
