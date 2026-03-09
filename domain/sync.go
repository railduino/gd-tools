package domain

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var SyncCommand = &cli.Command{
	Name:  "sync",
	Usage: "synchronize an email domain with the production server",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagForce,
		config.FlagSkipDNS,
		config.FlagSkipMX,
		&cli.BoolFlag{
			Name:  "add",
			Usage: "add new domain, update existing if not given",
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "update all existing domains",
		},
		&cli.BoolFlag{
			Name:  "dkim",
			Usage: "update DKIM signature for OpenDKIM",
		},
		&cli.StringFlag{
			Name:  "dmarc",
			Usage: "DMARC value (empty to use basics default)",
		},
	},
	ArgsUsage: "<domain> or --all",
	BashComplete: func(c *cli.Context) {
		dl, _, err := email.GetDomains(nil)
		if err != nil {
			return
		}
		for _, d := range dl.Domains {
			fmt.Fprintln(c.App.Writer, d.Name)
		}
	},
	Action: SyncRun,
}

func SyncRun(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	emailList, emailMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	var nameList []string
	if c.Bool("all") {
		if c.Bool("add") || c.NArg() != 0 {
			return fmt.Errorf("cannot use --all with --add or domain")
		}
		for _, domain := range emailList.Domains {
			nameList = append(nameList, domain.Name)
		}
	} else {
		if c.NArg() != 1 {
			return fmt.Errorf("missing domain to add or update")
		}
		nameList = append(nameList, c.Args().Get(0))
	}

	var updateList []*email.Domain
	for _, domainName := range nameList {
		domain, ok := emailMap[domainName]
		if ok {
			cfg.Sayf("✅ domain %s is managed by this server", domain.Name)
		} else if c.Bool("add") {
			cfg.Sayf("adding domain %s to this server", domainName)
			domain = &email.Domain{Name: domainName}
			emailList.Domains = append(emailList.Domains, domain)
			emailMap[domain.Name] = domain
			cfg.Force = true
		} else {
			return fmt.Errorf("domain %s neither found nor added", domainName)
		}
		if !cfg.Force && !c.Bool("all") {
			return nil
		}
		updateList = append(updateList, domain)

		if _, err := domain.EnsureLocalDKIM(c.Bool("dkim")); err != nil {
			return fmt.Errorf("failed to generate DKIM for %s: %w", domain.Name, err)
		}

		// There must always be a DMARC value
		if c.IsSet("dmarc") {
			domain.DMARC = c.String("dmarc")
		}
		if domain.DMARC == "" {
			domain.DMARC = cfg.DMARC
		}

		brevo, err := email.GetBrevo()
		if err != nil {
			return err
		}
		if brevo != nil && brevo.API_Key != "" {
			enabled, err := domain.BrevoUpdate(brevo.API_Key)
			if err != nil {
				return err
			}
			if !enabled {
				return fmt.Errorf("domain %s is missing in Brevo", domain.Name)
			}
		}

		if cfg.Spambarrier != "" {
			domain.AddSpamBarrier()
		} else {
			domain.MXs = []email.MX{
				{FQDN: cfg.FQDN(), Prio: 10},
			}
		}
	}

	if err := emailList.Save(); err != nil {
		return err
	}

	for _, domain := range updateList {
		if status, err := cfg.UpdateDomainDNS(domain); err != nil {
			return err
		} else if len(status) > 0 {
			cfg.Say(status)
		}
	}

	return nil
}
