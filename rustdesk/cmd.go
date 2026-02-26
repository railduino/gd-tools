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
	ArgsUsage: "<host>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 1 {
		return fmt.Errorf("Usage: gdt rustdesk <host>")
	}
	host := c.Args().Get(0)

	rdCfg := &agent.RustDesk{
		HostName: host,
	}

	// Prevent accidental collision with the main system FQDN
	if rdCfg.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for RustDesk")
	}

	// Persist via config layer (analog zu OCIS)
	if err := config.SaveRustDesk(rdCfg); err != nil {
		return err
	}

	return nil
}
