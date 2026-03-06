package basics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "basics",
	Usage: "Prepare or update gd-tools base directory",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "downloads",
			Usage: "update downloads.json from repository",
		},
		&cli.BoolFlag{
			Name:  "routing",
			Usage: "update routing.json from repository",
		},
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

	if _, err := os.Stat(agent.DownloadsName); err != nil || c.Bool("downloads") {
		tmpl := filepath.Join("assets", agent.DownloadsName)
		content, err := templates.Load(tmpl, false)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", agent.DownloadsName, err)
		}
		if err := os.WriteFile(agent.DownloadsName, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", agent.DownloadsName, err)
		}
		fmt.Println("created/updated downloads from repository")
	}

	if _, err := os.Stat(config.RoutingName); err != nil || c.Bool("routing") {
		tmpl := filepath.Join("assets", config.RoutingName)
		content, err := templates.Load(tmpl, false)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", config.RoutingName, err)
		}
		if err := os.WriteFile(config.RoutingName, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", config.RoutingName, err)
		}
		fmt.Println("created/updated routing from repository")
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

	if !basicsHasChanges(c) {
		printBasics(basics)
		return nil
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

func basicsHasChanges(c *cli.Context) bool {
	return c.String("company") != "" ||
		c.String("sysadmin") != "" ||
		c.String("help-url") != "" ||
		c.String("timezone") != "" ||
		c.String("language") != "" ||
		c.String("region") != "" ||
		c.Int("reg-ttl") != utils.DefaultRegTTL
}

func printBasics(b utils.Basics) {
	fmt.Printf("%-12s  %s\n", "KEY", "VALUE")
	fmt.Printf("%-12s  %s\n", "------------", "------------------------------")

	fmt.Printf("%-12s  %s\n", "Company", b.Company)
	fmt.Printf("%-12s  %s\n", "SysAdmin", b.SysAdmin)
	fmt.Printf("%-12s  %s\n", "HelpURL", b.HelpURL)
	fmt.Printf("%-12s  %s\n", "TimeZone", b.TimeZone)
	fmt.Printf("%-12s  %s\n", "Language", b.Language)
	fmt.Printf("%-12s  %s\n", "Region", b.Region)
	fmt.Printf("%-12s  %d\n", "RegTTL", b.RegTTL)
}
