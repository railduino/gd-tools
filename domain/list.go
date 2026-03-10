package domain

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name:  "list",
	Usage: "list existing Email/DNS domains (incl. accounts)",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	Action: ListRun,
}

func ListRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	domainList, _, err := email.GetDomains(nil)
	if err != nil {
		return err
	}
	if err := domainList.Save(); err != nil {
		return err
	}

	names := c.Args().Slice()
	if len(names) == 0 {
		for _, dom := range domainList.Domains {
			names = append(names, dom.Name)
		}
	}

	space := false
	for _, name := range names {
		for _, dom := range domainList.Domains {
			if dom.Name == name {
				if space {
					fmt.Println()
				}
				space = true

				lines, err := dom.Info()
				if err != nil {
					return err
				}
				for _, line := range lines {
					fmt.Println(line)
				}
				break
			}
		}
	}

	return nil
}
