package account

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var BrevoCommand = &cli.Command{
	Name:  "brevo",
	Usage: "setup Brevo for outbound emails (see https://www.brevo.com/de/)",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagDelete,
		config.FlagForce,
		&cli.StringFlag{
			Name:  "api",
			Usage: "API key",
		},
		&cli.StringFlag{
			Name:  "id",
			Usage: "SMTP identifier",
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "SMTP key",
		},
	},
	Action: BrevoRun,
}

func BrevoRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.Bool("delete") && c.Bool("force") {
		// TODO remove brevo.json if it exists
		return nil
	}

	brv, err := email.ReadBrevo(c)
	if err != nil {
		return err
	}

	if c.IsSet("api") {
		brv.API_Key = c.String("api")
	}
	if c.IsSet("id") {
		brv.SMTP_ID = c.String("id")
	}
	if c.IsSet("key") {
		brv.SMTP_Key = c.String("key")
	}

	fmt.Printf("Brevo: '%v'\n", brv)

	if err := brv.Save(); err != nil {
		return err
	}

	return nil
}
