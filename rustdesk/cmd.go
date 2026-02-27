package rustdesk

import (
	"fmt"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The rustdesk command prepares a RustDesk server instance for deployment.

RustDesk Server consists of:
- hbbs: ID/Rendezvous server
- hbbr: Relay server

This command prepares the configuration that will later be processed by deploy/system
logic to install binaries, configure systemd units, manage firewall ports, and
synchronize server keys via mTLS request/response.`

var Command = &cli.Command{
	Name:        "rustdesk",
	Usage:       "Prepare a RustDesk server instance for deployment",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
	},
	ArgsUsage: "<host> <domain>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		return fmt.Errorf("Usage: gdt rustdesk <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)

	rdCfg := &agent.RustDesk{
		HostName:   host,
		DomainName: domain,
	}

	// Prevent accidental collision with the main system FQDN
	if rdCfg.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for RustDesk")
	}

	if err := rdCfg.Save(); err != nil {
		return err
	}

	return nil
}
