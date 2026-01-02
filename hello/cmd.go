package hello

import (
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "hello",
	Usage: "Check mTLS communication with production server",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagPort,
		&cli.IntFlag{
			Name:  "timeout",
			Usage: "connection timeout in seconds (overrides config)",
			Value: 0, // 0 = use config
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, req, err := config.ReadConfigPlus(c)
	if err != nil {
		return err
	} else if cfg != nil {
		defer cfg.Close()
	}

	req.Hello = "Hi, there."

	if err := req.Send(); err != nil {
		return err
	}

	return nil
}
