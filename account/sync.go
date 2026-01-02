package account

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
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
			Name:  "dkim",
			Usage: "DKIM selector, defaults to gd-tools",
		},
		&cli.StringFlag{
			Name:  "dmarc",
			Usage: "default DMARC level: relaxed, *medium, strict",
			Value: "medium",
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
		return fmt.Errorf("domain %s neither found nor added", domain.Name)
	}
	if !cfg.Force {
		return nil
	}

	if dkim := c.String("dkim"); dkim != "" {
		domain.DKIM.Selector = dkim
	} else if domain.DKIM.Selector == "" {
		domain.DKIM.Selector = "gd-tools"
	}
	if domain.DKIM.PrivKey == "" || domain.DKIM.PubValue == "" {
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate private key for %s: %w", domain.Name, err)
		}
		privData := x509.MarshalPKCS1PrivateKey(privKey)
		privBlk := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privData}
		domain.DKIM.PrivKey = string(pem.EncodeToMemory(privBlk))

		der, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to marshal public key for %s: %w", domain.Name, err)
		}
		domain.DKIM.PubValue = base64.StdEncoding.EncodeToString(der)
	}

	domain.DMARC = c.String("dmarc")

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
