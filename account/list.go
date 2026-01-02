package account

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var ListCommand = &cli.Command{
	Name:  "list",
	Usage: "list existing email domains and accounts",
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

	for _, name := range names {
		for _, dom := range domainList.Domains {
			if dom.Name == name {
				fmt.Printf("Domain ........: %s\n", dom.Name)
				for _, user := range dom.UserList {
					fmt.Printf("  User ........: %s\n", user.Email())
				}
				fmt.Println()
				break
			}
		}
	}

	return nil
}
