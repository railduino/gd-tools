package domain

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var SPFCommand = &cli.Command{
	Name:  "spf",
	Usage: "add or delete SPF includes for domain",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	ArgsUsage: "<domain> add|delete <sender>",
	Action:    SPFRun,
}

func SPFRun(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 3 {
		return fmt.Errorf("[spf] missing action and/or sender value")
	}
	name := c.Args().Get(0)
	action := c.Args().Get(1)
	sender := c.Args().Get(2)

	domainList, domainMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	domain, ok := domainMap[name]
	if !ok {
		return fmt.Errorf("[spf] domain '%s' not found", name)
	}

	switch action {
	case "add":
		domain.AddSPF(sender)
	case "delete":
		domain.DeleteSPF(sender)
	default:
		return fmt.Errorf("[spf] invalid action '%s'", action)
	}

	if err := domainList.Save(); err != nil {
		return err
	}

	if status, err := cfg.UpdateDomainDNS(domain); err != nil {
		return fmt.Errorf("[spf] failed to update SPF: %v", err)
	} else if len(status) > 0 {
		cfg.Say(status)
	}

	return nil
}
