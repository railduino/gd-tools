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
		&cli.BoolFlag{
			Name:  "enabled",
			Usage: "activity status",
			Value: false,
		},
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
			Name:  "id",
			Usage: "identifier (user ID)",
		},
		&cli.StringFlag{
			Name:  "code",
			Usage: "Code for domain verification",
		},
		&cli.StringFlag{
			Name:  "key",
			Usage: "SMTP-Key for SASL authentication",
		},
		&cli.StringFlag{
			Name:  "dmarc",
			Usage: "proposed DMARC value",
			Value: "v=DMARC1; p=none; rua=mailto:rua@dmarc.brevo.com",
		},
	},
	Action: BrevoRun,
}

func BrevoRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	brevo, err := email.ReadBrevo(c)
	if err != nil {
		return err
	}

	updated := false
	if c.IsSet("enabled") {
		brevo.Enabled = c.Bool("enabled")
		updated = true
	}
	if c.IsSet("id") {
		brevo.ID = c.String("id")
		updated = true
	}
	if c.IsSet("code") {
		brevo.Code = c.String("code")
		updated = true
	}
	if c.IsSet("key") {
		brevo.Key = c.String("key")
		updated = true
	}

	if updated {
		if err := brevo.Save(); err != nil {
			return err
		}
	}

	return nil
}
