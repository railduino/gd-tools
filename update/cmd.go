package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/setup"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `Update selected parts of config.json for an existing host

This command is meant to be executed from within a host directory. It updates
local configuration on the development workstation and does not directly modify
the production server.

For detailed documentation and usage examples, see:
https://github.com/railduino/gd-tools/wiki/11-Update`

var Command = &cli.Command{
	Name:        "update",
	Usage:       "Update config.json content for a production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "basics",
			Usage: "update values from $(GD_TOOLS_BASE)/basics.json",
		},
		&cli.BoolFlag{
			Name:  "routing",
			Usage: "update routing.json from server root directory",
		},
		setup.FlagDMARC,
		setup.FlagCompany,
		setup.FlagSysAdmin,
		setup.FlagHelpURL,
		setup.FlagUbuntuPro,
		setup.FlagSpamBarrier,
		setup.FlagHetznerDNS,
		setup.FlagIonosDNS,
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	// Update from basics.json if requested (except company/sysadmin/help-url)
	if c.Bool("basics") {
		basics, err := utils.GetBasics()
		if err != nil {
			return err
		}
		cfg.TimeZone = basics.TimeZone
		cfg.Language = basics.Language
		cfg.Region = basics.Region
		cfg.RegTTL = basics.RegTTL
		cfg.DMARC = basics.DMARC
	}

	// Update routing.json if requested
	if c.Bool("routing") {
		path := filepath.Join("..", config.RoutingName)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", config.RoutingName, err)
		}
		if err := os.WriteFile(config.RoutingName, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", config.RoutingName, err)
		}
	}

	// There must always be a DMARC value
	if c.IsSet("dmarc") {
		cfg.DMARC = c.String("dmarc")
	}
	if cfg.DMARC == "" {
		cfg.DMARC = utils.DefaultDMARC
	}

	// Use IsSet so values can be explicitly cleared with --company "" etc.
	if c.IsSet("company") {
		cfg.Company = c.String("company")
	}
	if c.IsSet("sysadmin") {
		cfg.SysAdmin = c.String("sysadmin")
	}
	if c.IsSet("help-url") {
		cfg.HelpURL = c.String("help-url")
	}

	// Tokens / keys (also allow explicit clearing)
	if c.IsSet("spambarrier") {
		cfg.Spambarrier = c.String("spambarrier")
	}
	if c.IsSet("ubuntu-pro") {
		cfg.UbuntuPro = c.String("ubuntu-pro")
	}
	if c.IsSet("hetzner-dns") {
		cfg.HetznerToken = c.String("hetzner-dns")
	}
	if c.IsSet("ionos-dns") {
		cfg.IonosToken = c.String("ionos-dns")
	}

	return cfg.Save()
}
