package release

import "github.com/urfave/cli/v2"

var Describe = `The release command shows information about embedded application releases.`

var Command = &cli.Command{
	Name:        "release",
	Usage:       "Show information about embedded application releases",
	Description: Describe,
	Subcommands: []*cli.Command{
		InfoCommand,
	},
	Action: InfoRun,
}
