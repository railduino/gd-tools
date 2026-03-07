package wbce

import (
	"fmt"
	"sort"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/releases"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The wbce command prepares a WBCE CMS instance for deployment.

WBCE CMS is short for WebsiteBaker Community Edition Content Management System.
It was created in 2015 as a fork of WebsiteBaker, a popular but aging CMS.

Homepage ...: (en) https://wbce-cms.org/
Homepage ...: (de) https://wbce.org/de/wbce/
Releases ...: https://github.com/WBCE/WBCE_CMS/releases
Installation: (de) https://wbce.org/de/wbce/wbce_cms_installieren/

The <name> shall be a short name used to identify the project.
In case the host part is 'www' the domain itself will also be
registered with Lets Encrypt, and a non-WWW scheme will be used.`

var Command = &cli.Command{
	Name:        "wbce",
	Usage:       "Prepare a WBCE CMS instance for deployment",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "update",
			Usage: "update an existing WBCE CMS instance",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "password for MySQL access",
		},
		&cli.StringFlag{
			Name:  "salt",
			Usage: "base for various secret values",
		},
		&cli.StringFlag{
			Name:  "admin-name",
			Usage: "name of WBCE CMS admin",
			Value: "WBCE CMS Admin",
		},
		&cli.StringFlag{
			Name:  "admin-email",
			Usage: "email address of WBCE CMS admin",
		},
		&cli.StringFlag{
			Name:  "admin-pswd",
			Usage: "password of WBCE CMS admin",
		},
		&cli.StringSliceFlag{
			Name:  "alias",
			Usage: "alias domain (only if host is 'www')",
		},
	},
	ArgsUsage: "<host> <domain>",
	Action:    Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() != 2 {
		return fmt.Errorf("Usage: gdt wbce <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)
	aliases := c.StringSlice("alias")
	sort.Strings(aliases)

	cat, err := releases.Load(cfg.Verbose)
	if err != nil {
		return err
	}
	pr, rel, err := cat.Get("wbce", "")
	if err != nil {
		return err
	}

	entry := config.CMS{
		HostName:   host,
		DomainName: domain,
		Product:    "wbce",
		Version:    pr.Default,
		Language:   cfg.Language,
		Region:     cfg.Region,
		Password:   c.String("password"),
		DirName:    rel.Download.Directory,
		Salt:       c.String("salt"),
		AdminName:  c.String("admin-name"),
		AdminEmail: c.String("admin-email"),
		AdminPswd:  c.String("admin-pswd"),
		Aliases:    aliases,
	}
	if entry.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for WBCE CMS")
	}

	if c.Bool("update") {
		if _, err := config.LoadCMSList(&entry); err != nil {
			return err
		}
		cfg.Sayf("WBCE CMS '%s' was updated", entry.FQDN())
		return nil
	}

	list, err := config.LoadCMSList(nil)
	if err != nil {
		return err
	}

	entry.Number = config.FirstCMS
	for _, check := range list.Entries {
		if check.FQDN() == entry.FQDN() {
			return fmt.Errorf("the FQDN %s is already taken", entry.FQDN())
		}
		if check.Number == entry.Number {
			entry.Number = entry.Number + 1
			if entry.Number > config.LastCMS {
				return fmt.Errorf("no more WBCE CMS, please")
			}
		}
	}

	if entry.Password == "" {
		if entry.Password, err = utils.CreatePassword(20); err != nil {
			return err
		}
	}

	if entry.Salt = c.String("salt"); entry.Salt == "" {
		if entry.Salt, err = utils.CreatePassword(30); err != nil {
			return err
		}
	}

	adminEmail := c.String("admin-email")
	if adminEmail == "" {
		adminEmail = "admin@" + domain
	}
	adminUser, err := email.MakeUser(adminEmail)
	if err != nil {
		return err
	}
	entry.AdminEmail = adminUser.Email()

	secrets, err := utils.LoadSecrets()
	if err != nil {
		return err
	}
	_, _, err = secrets.SetMailUser(adminUser.Email(), "")
	if err != nil {
		return err
	}

	if entry.AdminPswd == "" {
		if entry.AdminPswd, err = utils.CreatePassword(20); err != nil {
			return err
		}
	}

	list.Entries = append(list.Entries, &entry)

	if err := list.Save(); err != nil {
		return err
	}

	return nil
}
