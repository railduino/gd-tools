package account

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var SyncCommand = &cli.Command{
	Name:  "sync",
	Usage: "add or update an email domain",
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
		&cli.StringFlag{
			Name:  "dmarc",
			Usage: "specific DMARC value (Brevo will override)",
		},
	},
	ArgsUsage: "<domain>",
	Action:    SyncRun,
}

func SyncRun(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 1 {
		return fmt.Errorf("missing domain to add or update")
	}
	domainName := c.Args().Get(0)

	emailList, emailMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

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
	if !cfg.Force {
		return nil
	}

	if _, err := domain.EnsureLocalDKIM(); err != nil {
		return fmt.Errorf("failed to generate DKIM for %s: %w", domain.Name, err)
	}

	if c.IsSet("dmarc") {
		domain.DMARC = c.String("dmarc")
	} else if domain.DMARC == "" {
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

	if err := emailList.Save(); err != nil {
		return err
	}

	if status, err := cfg.UpdateDomainDNS(domain); err != nil {
		return err
	} else if len(status) > 0 {
		cfg.Say(status)
	}

	return nil
}
