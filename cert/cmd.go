package cert

import (
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The cert command manages ACME certificates.

It scans project folders for "acme-certs/<domain>/" and operates on the found certificates.

Subcommands:
  - list   : Scan and show certificate status
  - delete : Delete a certificate directory (requires --force)
  - sync   : Sync certificates to production servers.`

var Command = &cli.Command{
	Name:        "cert",
	Usage:       "List, renew, delete or sync ACME certificates",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Subcommands: []*cli.Command{
		ListCommand,
		// TODO (later) DeleteCommand,
		// TODO (later) SyncCommand,
	},
	Action: ListRun,
}
