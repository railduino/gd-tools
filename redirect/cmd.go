package redirect

import (
	"fmt"
	"sort"
	"strings"

	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The redirect command sets up a redirect.

A redirect consists of a host and domain (the 'from' part) and a destination (the 'to' part).
If the host part is 'www' the domain itself will also be redirected to the target.

The destination will be prefixed with 'http(s)://' and may contain a complete URL path`

var Command = &cli.Command{
	Name:        "redirect",
	Usage:       "Setup a redirect from source to target",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringSliceFlag{
			Name:  "alias",
			Usage: "alias (redirected) domain (only if source starts with 'www')",
		},
	},
	ArgsUsage: "<host> <domain> <destination>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 3 {
		return fmt.Errorf("Usage: gdt redirect <host> <domain> <destination>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)
	destination := c.Args().Get(2)
	aliases := c.StringSlice("alias")
	sort.Strings(aliases)

	destination, _ = strings.CutPrefix(destination, "http://")
	destination, _ = strings.CutPrefix(destination, "https://")

	list, err := config.LoadRedirectList()
	if err != nil {
		return err
	}

	entry := &config.Redirect{
		HostName:    host,
		DomainName:  domain,
		Destination: destination,
		AdminEmail:  "admin@" + domain,
	}

	if host == "www" {
		entry.Aliases = aliases
	} else {
		return fmt.Errorf("aliases are only allowed if host is 'www'")
	}

	if entry.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for redirect")
	}

	found := false
	for index, check := range list.Entries {
		if check.FQDN() == entry.FQDN() {
			list.Entries[index] = entry
			found = true
			break
		}
	}

	if !found {
		list.Entries = append(list.Entries, entry)
	}

	if err := list.Save(); err != nil {
		return err
	}

	return nil
}
