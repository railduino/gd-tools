package nextcloud

import (
	"crypto/rand"
	"fmt"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The nextcloud command prepares a Nextcloud instance for deployment.

Nextcloud is a self-hosted platform for file sharing, collaboration, and cloud services.

The details will be elaborated when the command is stable.`

var Command = &cli.Command{
	Name:        "nextcloud",
	Usage:       "Prepare a Nextcloud instance for deployment",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "update",
			Usage: "update an existing Nextcloud instance",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "password for MySQL access",
		},
		&cli.StringFlag{
			Name:  "instance",
			Usage: "instance id to be installed",
		},
		&cli.StringFlag{
			Name:  "salt",
			Usage: "password hashing salt",
		},
		&cli.StringFlag{
			Name:  "secret",
			Usage: "base for secret key and auth values",
		},
		&cli.StringFlag{
			Name:  "admin",
			Usage: "email address of nextcloud admin",
		},
		&cli.StringFlag{
			Name:  "subdir",
			Usage: "subdirectory if not running from root",
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
		return fmt.Errorf("Usage: gdt nextcloud <host> <domain>")
	}
	host := c.Args().Get(0)
	domain := c.Args().Get(1)

	entry := agent.Nextcloud{
		HostName:   host,
		DomainName: domain,
		Subdir:     c.String("subdir"),
		Language:   cfg.Language,
		Region:     cfg.Region,
		Password:   c.String("password"),
		InstanceID: c.String("instance"),
		Salt:       c.String("salt"),
		Secret:     c.String("secret"),
		AdminEmail: c.String("admin"),
		AppList:    []string{},
	}
	if entry.FQDN() == cfg.FQDN() {
		return fmt.Errorf("cannot use the server name for Nextcloud")
	}
	entry.Name = agent.MakeDBName(entry.FQDN())

	if c.Bool("update") {
		if _, err := agent.LoadNextcloudList(&entry); err != nil {
			return err
		}
		cfg.Sayf("Nextcloud '%s' was updated", entry.Name)
		return nil
	}

	list, err := agent.LoadNextcloudList(nil)
	if err != nil {
		return err
	}

	entry.Number = agent.FirstNextcloud
	for _, check := range list.Entries {
		if check.Name == entry.Name {
			return fmt.Errorf("the name %s is already taken", entry.Name)
		}
		if check.FQDN() == entry.FQDN() {
			return fmt.Errorf("the FQDN %s is already taken", entry.FQDN())
		}
		if check.Number == entry.Number {
			if entry.Number = entry.Number + 1; entry.Number >= agent.LastNextcloud {
				return fmt.Errorf("no more Nextcloud, please")
			}
		}
	}

	if entry.InstanceID == "" {
		if entry.InstanceID, err = GenerateInstanceID(); err != nil {
			return err
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

	if entry.Secret = c.String("secret"); entry.Secret == "" {
		if entry.Secret, err = utils.CreatePassword(48); err != nil {
			return err
		}
	}

	admin := c.String("admin")
	if admin == "" {
		admin = fmt.Sprintf("Nextcloud Admin <nextcloud@%s>", cfg.DomainName)
	}
	adminUser, err := email.MakeUser(admin)
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

	list.Entries = append(list.Entries, &entry)

	if err := list.Save(); err != nil {
		return err
	}

	return nil
}

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

func GenerateInstanceID() (string, error) {
	const length = 10
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}

	for i, b := range buf {
		buf[i] = chars[int(b)%len(chars)]
	}

	return string(buf), nil
}
