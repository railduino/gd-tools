package ocis

import (
	"fmt"
	"os"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The ocis command prepares an OCIS instance for deployment.

OCIS (ownCloud Infinite Scale) is a new way of storing and sharing files.
Documentation is available at https://owncloud.dev/ocis/getting-started/

The details will be elaborated when the command is stable.`

var Command = &cli.Command{
	Name:        "ocis",
	Usage:       "Prepare an OCIS instance for deployment",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringFlag{
			Name:  "admin",
			Usage: "name of OCIS admin user",
			Value: "admin",
		},
		&cli.StringFlag{
			Name:  "email",
			Usage: "email of OCIS admin",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "password for OCIS admin",
		},
	},
	ArgsUsage: "<host> <domain>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if info, err := os.Stat(config.OCISName); err == nil && info.Size() > 0 {
		return fmt.Errorf("only one OCIS instance is allowed")
	}

	if c.NArg() != 2 {
		return fmt.Errorf("Usage: gdt ocis <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)

	ocCfg := config.OCIS{
		HostName:   host,
		DomainName: domain,
		AdminName:  c.String("admin"),
		Language:   cfg.Language,
		LogLevel:   "info",
	}
	if ocCfg.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for OCIS")
	}

	if ocCfg.AdminEmail = c.String("email"); ocCfg.AdminEmail == "" {
		ocCfg.AdminEmail = cfg.SysAdmin
	}

	if ocCfg.Password = c.String("password"); ocCfg.Password == "" {
		ocCfg.Password, err = utils.CreatePassword(20)
		if err != nil {
			return err
		}
	}

	if err := ocCfg.Save(); err != nil {
		return err
	}

	return nil
}
