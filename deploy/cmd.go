package deploy

import (
	"fmt"
	"os"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Describe = `The deploy command is used to update the production system.

Valid arguments are:
  - 00-system (includes bootstrap ... unbound)
  - 01-bootstrap
  - 02-packages
  - 03-filesystem
  - 04-database (Mariadb and redis)
  - 05-php      (currently 8.3)
  - 06-apache   (with ACME certs)
  - 07-redirect
  - 08-unbound  (local DNS resolver)

  - 10-email (includes dovecot ... roundcube)
  - 11-postfix
  - 12-dovecot
  - 13-rspamd
  - 14-accounts
  - 15-roundcube

  - 20      nextcloud (all)
  - 21...24 nextcloud (individual)
  - 25      ocis      (ownCloud Infinite Scale)

  - 30      CMS systems (all)
            WordPress, WBCE CMS, CMS Made Simple, etc.
  - 31...39 CMS system (individual)

  - 41-rustdesk

  - 90-finish (includes backup ... fail2ban)
  - 91-backup (Borg)
  - 92-fail2ban

  - 99-all (includes everything)`

var Command = &cli.Command{
	Name:        "deploy",
	Usage:       "Deploy some or all system and/or project components",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagForce,
		config.FlagPort,
		config.FlagSkipDNS,
		config.FlagSkipMX,
		&cli.BoolFlag{
			Name:  "upgrade",
			Usage: "update the system (apt-get update/upgrade)",
		},
	},
	ArgsUsage: "[<component>]...",
	BashComplete: func(c *cli.Context) {
		fmt.Println("00-system")
		fmt.Println("01-bootstrap")
		fmt.Println("02-packages")
		fmt.Println("03-filesystem")
		fmt.Println("04-database")
		fmt.Println("05-php")
		fmt.Println("06-apache")
		fmt.Println("07-redirect")
		fmt.Println("08-unbound")

		fmt.Println("10-email")
		fmt.Println("11-postfix")
		fmt.Println("12-dovecot")
		fmt.Println("13-rspamd")
		fmt.Println("14-accounts")
		fmt.Println("15-roundcube")

		fmt.Println("20-nextcloud")
		ncList, err := agent.LoadNextcloudList(nil)
		if err == nil && len(ncList.Entries) > 0 {
			for _, nc := range ncList.Entries {
				fmt.Printf("%02d-%s\n", nc.Number, nc.FQDN())
			}
		}

		if info, err := os.Stat(config.OCISName); err == nil && info.Size() > 0 {
			fmt.Println("25-ocis")
		}

		cmsList, err := config.LoadCMSList(nil)
		if err == nil && len(cmsList.Entries) > 0 {
			fmt.Println("30-cms")
			for _, cms := range cmsList.Entries {
				fmt.Printf("%02d-%s\n", cms.Number, cms.FQDN())
			}
		}

		if info, err := os.Stat(agent.RustDeskName); err == nil && info.Size() > 0 {
			fmt.Println("41-rustdesk")
		}

		fmt.Println("90-finish")
		fmt.Println("91-backup")
		fmt.Println("92-fail2ban")
		fmt.Println("99-all")
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, _, err := config.ReadConfigPlus(c)
	if err != nil {
		return err
	} else if cfg != nil {
		defer cfg.Close()
	}

	if err := cfg.SetupCA(); err != nil {
		return err
	}

	args := c.Args().Slice()

	if ShouldRun(args, "00-system", "00", "01-bootstrap", "01") {
		if err := cfg.DeployBootstrap(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "02-packages", "02") {
		upgrade := c.Bool("upgrade")
		if err := cfg.DeployPackages(upgrade); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "03-filesystem", "03") {
		if err := cfg.DeployMounts(); err != nil {
			return err
		}
		if err := cfg.DeployFilesystem(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "04-database", "04") {
		if err := cfg.DeployDatabase(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "05-php", "05") {
		if err := cfg.DeployPHP(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "06-apache", "06") {
		if err := cfg.DeployApache(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "07-redirect", "07") {
		if err := cfg.DeployRedirect(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "00-system", "00", "08-unbound", "08") {
		if err := cfg.DeployUnbound(); err != nil {
			return err
		}
	}

	if ShouldRun(args, "10-email", "10", "11-postfix", "11") {
		if err := cfg.DeployPostfix(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "10-email", "10", "12-dovecot", "12") {
		if err := cfg.DeployDovecot(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "10-email", "10", "13-rspamd", "13") {
		if err := cfg.DeployRspamd(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "10-email", "10", "14-accounts", "14") {
		if err := cfg.DeployAccounts(); err != nil {
			return err
		}
	}
	if ShouldRun(args, "10-email", "10", "15-roundcube", "15") {
		if err := cfg.DeployRoundcube(); err != nil {
			return err
		}
	}

	ncList, err := agent.LoadNextcloudList(nil)
	if err != nil {
		return err
	}
	for _, nc := range ncList.Entries {
		fqdn := fmt.Sprintf("%02d-%s", nc.Number, nc.FQDN())
		num := fmt.Sprintf("%02d", nc.Number)
		if ShouldRun(args, "20-cloud", "20", fqdn, num) {
			cfg.Nextcloud = nc
			if err := cfg.DeployNextcloud(); err != nil {
				return err
			}
		}
	}

	if info, err := os.Stat(config.OCISName); err == nil && info.Size() > 0 {
		if ShouldRun(args, "25-ocis", "25") {
			if err := cfg.DeployOCIS(); err != nil {
				return err
			}
		}
	}

	cmsList, err := config.LoadCMSList(nil)
	if err != nil {
		return err
	}
	for _, cms := range cmsList.Entries {
		fqdn := fmt.Sprintf("%02d-%s", cms.Number, cms.FQDN())
		num := fmt.Sprintf("%02d", cms.Number)
		if ShouldRun(args, "30-wordpress", "30", fqdn, num) {
			cfg.CMS = cms
			if err := cfg.DeployCMS(); err != nil {
				return err
			}
		}
	}

	if info, err := os.Stat(agent.RustDeskName); err == nil && info.Size() > 0 {
		if ShouldRun(args, "41-rustdesk", "41") {
			if err := cfg.DeployRustDesk(); err != nil {
				return err
			}
		}
	}

	if ShouldRun(args, "90-finish", "90", "91-backup", "91") {
		if err := cfg.DeployBackup(); err != nil {
			return err
		}
	}

	// TODO (later) maybe add fail2ban

	return nil
}

func ShouldRun(args []string, names ...string) bool {
	for _, arg := range args {
		for _, name := range names {
			if arg == name || arg == "99" || arg == "99-all" {
				return true
			}
		}
	}
	return false
}
