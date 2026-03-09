package domain

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var AliasCommand = &cli.Command{
	Name:  "alias",
	Usage: "add or delete alias names for domain",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	ArgsUsage: "<domain> add|delete <alias>",
	Action:    AliasRun,
}

func AliasRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 3 {
		return fmt.Errorf("[alias] missing action and/or alias")
	}
	name := c.Args().Get(0)
	action := c.Args().Get(1)
	alias := c.Args().Get(2)

	domainList, domainMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	domain, ok := domainMap[name]
	if !ok {
		return fmt.Errorf("[alias] domain '%s' not found", name)
	}

	switch action {
	case "add":
		domain.AddAlias(alias)
	case "delete":
		domain.DeleteAlias(alias)
	default:
		return fmt.Errorf("[alias] invalid action '%s'", action)
	}

	if err := domainList.Save(); err != nil {
		return err
	}

	return nil
}
