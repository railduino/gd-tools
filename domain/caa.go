package domain

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var CAACommand = &cli.Command{
	Name:  "caa",
	Usage: "add or delete CAA records for domain",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	ArgsUsage: "<domain> add|delete <caa>",
	Action:    CAARun,
}

func CAARun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	domainList, domainMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	name := ""
	switch c.NArg() {
	case 0: // show all domains
	case 1:
		name = c.Args().Get(0)
		if _, ok := domainMap[name]; !ok {
			return fmt.Errorf("[caa] domain '%s' not found", name)
		}
	case 3:
		name = c.Args().Get(0)
		dom, ok := domainMap[name]
		if !ok {
			return fmt.Errorf("[caa] domain '%s' not found", name)
		}
		switch action := c.Args().Get(1); action {
		case "add":
			dom.AddCAA(c.Args().Get(2))
		case "delete":
			dom.DeleteCAA(c.Args().Get(2))
		default:
			return fmt.Errorf("[caa] invalid action '%s'", action)
		}
	default:
		return fmt.Errorf("[caa] missing action and/or caa")
	}

	if err := domainList.Save(); err != nil {
		return err
	}

	space := false
	for _, dom := range domainList.Domains {
		if name != "" && name != dom.Name {
			continue
		}

		if space {
			fmt.Println("")
		}
		space = true

		fmt.Printf("Domain: %s\n", dom.Name)
		for _, caa := range dom.CAAs {
			fmt.Printf("   CAA: %s\n", caa)
		}
	}

	return nil
}
