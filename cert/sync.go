package cert

import (
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var SyncCommand = &cli.Command{
	Name:   "sync",
	Usage:  "Sync active ACME certificates to the production server",
	Action: SyncRun,
}

func SyncRun(c *cli.Context) error {
	if err := utils.EnsureHostDir(); err != nil {
		return err
	}

	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	// PushCerts is intentionally strict:
	// - rsync --delete is always enabled
	// - errors are logged but not fatal
	cfg.PushCerts()

	return nil
}
