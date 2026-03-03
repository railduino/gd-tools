package account

import (
	// "fmt"

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
			Name:  "server",
			Usage: "SMTP server",
			Value: "smtp-relay.brevo.com",
		},
		&cli.IntFlag{
			Name:  "port",
			Usage: "SMTP port",
			Value: 587,
		},
		&cli.StringFlag{
			Name:  "api (admin access)",
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

	brevo, err := email.ReadBrevo(c)
	if err != nil {
		return err
	}

	updated := false
	if c.IsSet("server") {
		brevo.Server = c.String("server")
		updated = true
	}
	if c.IsSet("port") {
		brevo.Port = c.Int("port")
		updated = true
	}
	if c.IsSet("api") {
		brevo.API_Key = c.String("api")
		updated = true
	}
	if c.IsSet("id") {
		brevo.SMTP_ID = c.String("id")
		updated = true
	}
	if c.IsSet("key") {
		brevo.SMTP_Key = c.String("key")
		updated = true
	}

	if updated {
		if err := brevo.Save(); err != nil {
			return err
		}
	}

	return nil
}
