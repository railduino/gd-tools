package sync

import (
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The sync command synchronizes the ACME certs.

From time to time it should be called; it does no harm anyway.`

var Command = &cli.Command{
	Name:        "sync",
	Usage:       "Sync ACME certs with the production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	cfg.PushCerts()

	return nil
}
