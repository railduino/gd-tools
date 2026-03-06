package account

import (
	"fmt"
	"strings"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var DeployCommand = &cli.Command{
	Name:  "deploy",
	Usage: "update account data on the production server",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagPort,
		config.FlagSkipDNS,
		config.FlagSkipMX,
		&cli.BoolFlag{
			Name:  "webmail",
			Usage: "deploy webmail in addition to proxies",
		},
		&cli.BoolFlag{
			Name:  "maintenance",
			Usage: "reverse user passwords for admin access",
		},
	},
	ArgsUsage: "[<domain> ...]",
	BashComplete: func(c *cli.Context) {
		dl, _, err := email.GetDomains(nil)
		if err != nil {
			return
		}
		for _, d := range dl.Domains {
			fmt.Fprintln(c.App.Writer, d.Name)
			for _, a := range d.Aliases {
				fmt.Fprintln(c.App.Writer, a)
			}
		}
	},
	Action: DeployRun,
}

func DeployRun(c *cli.Context) error {
	cfg, _, err := config.ReadConfigPlus(c)
	if err != nil {
		return err
	} else if cfg != nil {
		defer cfg.Close()
	}

	sel := make(map[string]bool)
	for _, a := range c.Args().Slice() {
		a = strings.TrimSpace(a)
		if a != "" {
			sel[a] = true
		}
	}

	if err := cfg.DeployAccountMap(sel); err != nil {
		return err
	}

	if c.Bool("webmail") {
		if err := cfg.DeployRoundcubeMap(sel); err != nil {
			return err
		}
	}

	return nil
}
