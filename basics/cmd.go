package basics

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "basics",
	Usage: "Prepare or update gd-tools base directory",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringFlag{
			Name:  "company",
			Usage: "Name of your organisation",
		},
		&cli.StringFlag{
			Name:  "sysadmin",
			Usage: "System administrator email",
		},
		&cli.StringFlag{
			Name:  "help-url",
			Usage: "support URL, e.g. for Webmail",
		},
		&cli.StringFlag{
			Name:  "timezone",
			Usage: "Timezone, e.g. Europe/Berlin",
		},
		&cli.StringFlag{
			Name:  "language",
			Usage: "Language, e.g. de",
		},
		&cli.StringFlag{
			Name:  "region",
			Usage: "Region, e.g. DE",
		},
		&cli.IntFlag{
			Name:  "reg-ttl",
			Usage: "Regular Time-To-Live, cache-time for DNS",
			Value: utils.DefaultRegTTL,
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	if err := utils.EnsureBaseDir(); err != nil {
		return err
	}

	var basics utils.Basics
	content, err := os.ReadFile(utils.BasicsName)
	if err != nil {
		if os.IsNotExist(err) {
			basics = utils.Basics{
				Company:  utils.DefaultCompany,
				SysAdmin: utils.GetSysAdmin(),
				HelpURL:  "mailto:" + utils.GetSysAdmin(),
				TimeZone: utils.GetTimeZone(),
				Language: agent.GetLanguage(),
				Region:   agent.GetRegion(),
				RegTTL:   utils.DefaultRegTTL,
			}
		} else {
			return fmt.Errorf("failed to read %s: %w", utils.BasicsName, err)
		}
	} else {
		if err := json.Unmarshal(content, &basics); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", utils.BasicsName, err)
		}
	}

	if value := c.String("company"); value != "" {
		basics.Company = value
	}
	if value := c.String("sysadmin"); value != "" {
		basics.SysAdmin = value
	}
	if value := c.String("help-url"); value != "" {
		basics.HelpURL = value
	}
	if value := c.String("timezone"); value != "" {
		basics.TimeZone = value
	}
	if value := c.String("language"); value != "" {
		basics.Language = value
	}
	if value := c.String("region"); value != "" {
		basics.Region = value
	}
	if secs := c.Int("reg-ttl"); secs != utils.DefaultRegTTL {
		basics.RegTTL = secs
	}

	if err := basics.Save(); err != nil {
		return err
	}

	return nil
}
