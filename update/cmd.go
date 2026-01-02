package update

import (
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The update command changes various config.json parts.

The details will be elaborated when the command is stable.`

var Command = &cli.Command{
	Name:        "update",
	Usage:       "Update config.json content for a production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "basics",
			Usage: "update values from basics.json",
		},
		&cli.StringFlag{
			Name:  "company",
			Usage: "Company name, used e.g. for Webmail",
		},
		&cli.StringFlag{
			Name:  "help-url",
			Usage: "Support URL for this server",
		},
		// TODO add Hetzner/IONOS/whatever API key
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	basics, err := utils.GetBasics()
	if err != nil {
		return err
	}

	if c.Bool("basics") {
		cfg.TimeZone = basics.TimeZone
		cfg.Language = basics.Language
		cfg.Region = basics.Region
		cfg.RegTTL = basics.RegTTL
		cfg.SysAdmin = basics.SysAdmin
	}

	if company := c.String("company"); company != "" {
		cfg.Company = company
	}

	if help := c.String("help-url"); help != "" {
		cfg.HelpURL = help
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	return nil
}
