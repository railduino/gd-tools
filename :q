package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/php"
	"github.com/railduino/gd-tools/templates"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

const (
	Version = "1.0"
)

var Describe = `Initialize a new production server

Creates the initial configuration on the development workstation.
Does not modify the production server.

For detailed documentation and usage examples, see:
https://github.com/railduino/gd-tools/wiki/10-Setup`

var Command = &cli.Command{
	Name:        "setup",
	Usage:       "Initialize a new production server",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringFlag{
			Name:  "hetzner-volume",
			Usage: "add a Hetzner Cloud Volume for /var/gd-tools",
		},
		&cli.StringFlag{
			Name:  "raid-device",
			Usage: "add a /dev/mdX RAID device for /var/gd-tools",
		},
		&cli.StringFlag{
			Name:  "swap-size",
			Usage: "e.g. '4G' - create or verify swapfile",
			Value: "0",
		},
		&cli.StringFlag{
			Name:  "dmarc",
			Usage: "default DMARC level: relaxed, strict, (any = medium)",
			Value: "medium",
		},
		&cli.StringFlag{
			Name:  "company",
			Usage: "Company name, used e.g. for Webmail",
		},
		&cli.StringFlag{
			Name:  "help-url",
			Usage: "Support URL for this server",
		},
		&cli.StringFlag{
			Name:  "hetzner-dns",
			Usage: "configure Hetzner Cloud DNS API token for declarative DNS management",
		},
		&cli.StringFlag{
			Name:  "ionos-dns",
			Usage: "configure IONOS DNS API token for declarative DNS management",
		},
		&cli.StringFlag{
			Name:  "ubuntu-pro",
			Usage: "attach Ubuntu Pro subscription using the provided token",
		},
		&cli.StringFlag{
			Name:  "spambarrier",
			Usage: "SpamBarrier API key for inbound email",
		},
		&cli.StringFlag{
			Name:  "brevo-code",
			Usage: "Brevo Code for for outbound email (domain verification)",
		},
		&cli.StringFlag{
			Name:  "brevo-key",
			Usage: "Brevo SMTP-Key for for outbound email (sasl authentication)",
		},
	},
	ArgsUsage: "<host> <domain>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	basics, err := utils.ReadBasics()
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		return fmt.Errorf("expected arguments: <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)

	cfg := config.Config{
		Version:      Version,
		Verbose:      c.Bool("verbose"),
		Dry:          c.Bool("dry"),
		TimeZone:     basics.TimeZone,
		Language:     basics.Language,
		Region:       basics.Region,
		RegTTL:       basics.RegTTL,
		HostName:     host,
		DomainName:   domain,
		DMARC:        c.String("dmarc"),
		SwapSize:     c.String("swap-size"),
		SysAdmin:     basics.SysAdmin,
		Company:      c.String("company"),
		HelpURL:      c.String("help-url"),
		Spambarrier:  c.String("spambarrier"),
		BrevoCode:    c.String("brevo-code"),
		BrevoKey:     c.String("brevo-key"),
		UbuntuPro:    c.String("ubuntu-pro"),
		HetznerToken: c.String("hetzner-dns"),
		IonosToken:   c.String("ionos-dns"),
	}

	fqdn := cfg.FQDN()
	configPath := filepath.Join(fqdn, config.ConfigName)

	if cfg.DMARC != "relaxed" && cfg.DMARC != "strict" {
		cfg.DMARC = "medium"
	}

	if cfg.Company == "" {
		cfg.Company = basics.Company
	}

	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("Server %s exists - will not overwrite", fqdn)
	}

	reservedNames := []string{
		"autoconfig",
		"autodiscover",
		"mta-sts",
		"imap",
		"smtp",
		"vmail",
		"webmail",
		"www",
	}
	for _, name := range reservedNames {
		if host == name {
			return fmt.Errorf("hostname '%s' is reserved", name)
		}
	}

	// collect default DEB packages
	packagesPath := filepath.Join("assets", "packages.txt")
	cfg.Packages, err = templates.Lines(packagesPath, "#", cfg.Verbose, php.GetTemplateData())
	if err != nil {
		return err
	}

	// read default downloads and known_hosts
	downloads, err := os.ReadFile(agent.DownloadsName)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", agent.DownloadsName, err)
	}
	khContent, khErr := os.ReadFile("known_hosts")

	// check for filesystems to be mounted
	// N.B. mounts given here are mutually exclusive
	if volume := c.String("hetzner-volume"); volume != "" {
		mount := agent.Mount{
			Provider: "Hetzner",
			ID:       volume,
			Dir:      agent.GetToolsDir(""),
		}
		cfg.Mounts = append(cfg.Mounts, &mount)
	} else if device := c.String("raid-device"); device != "" {
		mount := agent.Mount{
			Provider: "RAID",
			ID:       device,
			Dir:      agent.GetToolsDir(""),
		}
		cfg.Mounts = append(cfg.Mounts, &mount)
	}

	if cfg.Dry {
		cfg2 := cfg
		cfg2.Spambarrier = "***"
		cfg2.BrevoCode = "***"
		cfg2.BrevoKey = "***"
		cfg2.UbuntuPro = "***"
		cfg2.HetznerToken = "***"
		cfg2.IonosToken = "***"

		content, err := json.MarshalIndent(cfg2, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal %s: %w", fqdn, err)
		}
		cfg.Sayf("Config: >%s<", string(content))

		return nil
	}

	// From here on, we operate inside the host directory on the development workstation.
	if err := os.Mkdir(fqdn, 0755); err != nil {
		return err
	}
	if err := os.Chdir(fqdn); err != nil {
		return err
	}

	if err := cfg.SetupCA(); err != nil {
		return err
	}

	if err := os.WriteFile(agent.DownloadsName, downloads, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", agent.DownloadsName, err)
	}

	if khErr == nil {
		if err := os.WriteFile("known_hosts", khContent, 0600); err != nil {
			return fmt.Errorf("failed to write known_hosts: %w", err)
		}
	}

	if _, _, err := utils.GetRSAKeyPair(fqdn); err != nil {
		return err
	}

	if err := cfg.Save(); err != nil {
		return err
	}

	if err := os.MkdirAll(config.ACME_Cert_Dir, 0755); err != nil {
		return err
	}

	return nil
}
