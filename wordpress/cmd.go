package wordpress

import (
	"fmt"
	"sort"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The wordpress command prepares a WordPress instance for deployment.

WordPress is a free and open source blogging tool and
content management system (CMS) based on PHP and MySQL.

Homepage ...: https://wordpress.org/
Releases ...: https://wordpress.org/download/releases/
Installation: https://developer.wordpress.org/advanced-administration/before-install/howto-install/

In case the host part is 'www' the domain itself will also be
registered with Lets Encrypt, and a non-WWW scheme will be used.`

var Command = &cli.Command{
	Name:        "wordpress",
	Usage:       "Prepare a WordPress instance for deployment",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "update",
			Usage: "update an existing WordPress instance",
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
			Usage: "name of WordPress admin",
			Value: "WordPress Admin",
		},
		&cli.StringFlag{
			Name:  "admin-email",
			Usage: "email address of WordPress admin",
		},
		&cli.StringFlag{
			Name:  "admin-pswd",
			Usage: "password of WordPress admin",
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
		return fmt.Errorf("Usage: gdt wordpress <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)
	aliases := c.StringSlice("alias")
	sort.Strings(aliases)

	download, err := agent.GetDownload("wordpress")
	if err != nil {
		return err
	}
	if download.DirName == "" {
		return fmt.Errorf("missing DirName in wordpress download")
	}

	entry := config.CMS{
		HostName:   host,
		DomainName: domain,
		Product:    "wordpress",
		Language:   cfg.Language,
		Region:     cfg.Region,
		Password:   c.String("password"),
		DirName:    download.DirName,
		Salt:       c.String("salt"),
		AdminName:  c.String("admin-name"),
		AdminEmail: c.String("admin-email"),
		AdminPswd:  c.String("admin-pswd"),
		Aliases:    aliases,
	}
	if entry.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for WordPress")
	}

	if c.Bool("update") {
		if _, err := config.LoadCMSList(&entry); err != nil {
			return err
		}
		cfg.Sayf("WordPress '%s' was updated", entry.FQDN())
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
				return fmt.Errorf("no more WordPress, please")
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
