package sync

import (
	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The sync command synchronizes the ACME certs.

From time to time it should be called; it does no harm anyway.`

var Command = &cli.Command{
	Name:        "sync",
	Usage:       "Sync ACME certs with the production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		&cli.BoolFlag{
			Name:  "ssh-key",
			Usage: "also copy the ssh key pair",
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, req, err := config.ReadConfigPlus(c)
	if err != nil {
		return err
	}

	cfg.PushCerts()

	if c.Bool("ssh-key") {
		privContent, pubContent, err := utils.GetRSAKeyPair(cfg.FQDN())
		if err != nil {
			return err
		}
		privFile := agent.File{
			Task:    "write",
			Path:    agent.GetRootDir(".ssh", "id_rsa"),
			Content: privContent,
			Mode:    "0600",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&privFile)
		pubFile := agent.File{
			Task:    "write",
			Path:    agent.GetRootDir(".ssh", "id_rsa.pub"),
			Content: pubContent,
			Mode:    "0600",
			User:    "root",
			Group:   "root",
		}
		req.AddFile(&pubFile)

		if err := req.Send(); err != nil {
			return err
		}
	}

	return nil
}
