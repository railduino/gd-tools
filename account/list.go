package account

import (
	"fmt"
	"strings"

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

	brevo, err := email.GetBrevo()
	if err != nil {
		return err
	}
	var brevoMissing []string

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
				status := ""
				valid := false
				if brevo != nil && brevo.API_Key != "" {
					status, valid, err = dom.BrevoStatus(brevo.API_Key)
					if err != nil {
						return err
					}
					if !valid {
						brevoMissing = append(brevoMissing, dom.Name)
					}
				}
				fmt.Printf("Domain ........: %-24s%s\n", dom.Name, status)
				for _, user := range dom.UserList {
					fmt.Printf("  User ........: %s\n", user.Email())
				}
				fmt.Println()
				break
			}
		}
	}

	if len(brevoMissing) > 0 {
		fmt.Printf("################ Missing in Brevo: %s\n", strings.Join(brevoMissing, ", "))
	}

	return nil
}
