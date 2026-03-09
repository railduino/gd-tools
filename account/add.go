package account

import (
	"fmt"
	"slices"
	"strings"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var AddDescribe = `The 'account add' command adds or updates a mail user.

The email address can take several forms:
  - the plain address        jane.doe@example.com
  - the address in brackets  <john.doe@example.org>
  - the name and address     The Does <family.doe@example.net>

The brackets must be escaped from shell expansion, of course.`

var AddCommand = &cli.Command{
	Name:        "add",
	Usage:       "add a new email account to an existing domain",
	Description: AddDescribe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringFlag{
			Name:  "password",
			Usage: "system password (default: random string)",
		},
		&cli.StringFlag{
			Name:  "quota",
			Usage: "mailbox quota (default: unlimited)",
		},
		&cli.BoolFlag{
			Name:  "locked",
			Usage: "lock the account",
		},
		&cli.StringSliceFlag{
			Name:  "alias",
			Usage: "additional email alias for this user",
		},
	},
	ArgsUsage: "user@domain -or- 'Name <user@domain>'",
	Action:    AddRun,
}

func AddRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() < 1 {
		return fmt.Errorf("Usage: gdt account add [options] [name] user@domain")
	}
	addr := strings.Join(c.Args().Slice(), " ")

	newUser, err := email.MakeUser(addr)
	if err != nil {
		return fmt.Errorf("invalid email address '%s': %w", addr, err)
	}
	newUser.Quota = c.String("quota")
	newUser.Aliases = c.StringSlice("alias")
	newUser.Locked = c.Bool("locked")
	password := c.String("password")

	domainList, domainMap, err := email.GetDomains(nil)
	if err != nil {
		return fmt.Errorf("failed to load domains: %w", err)
	}

	domain, ok := domainMap[newUser.Domain]
	if !ok {
		return fmt.Errorf("invalid or unknown domain %s", newUser.Domain)
	}

	changed := false
	result := "added"
	oldUser, ok := domain.UserMap[newUser.Email()]
	if ok {
		if newUser.Name != "" && newUser.Name != oldUser.Name {
			oldUser.Name = newUser.Name
			changed = true
		}
		result = "updated"
	} else {
		oldUser, _ = email.MakeUser(newUser.Address())
		domain.UserList = append(domain.UserList, oldUser)
		changed = true
	}

	if newUser.Quota != oldUser.Quota {
		oldUser.Quota = newUser.Quota
		changed = true
	}

	if !slices.Equal(newUser.Aliases, oldUser.Aliases) {
		oldUser.Aliases = newUser.Aliases
		changed = true
	}

	if newUser.Locked != oldUser.Locked {
		oldUser.Locked = newUser.Locked
		changed = true
	}

	secrets, err := utils.LoadSecrets()
	if err != nil {
		return err
	}
	oldSecret := secrets.Get(utils.MailUserName, oldUser.Email())

	if password != "" || oldSecret == nil || oldSecret.Output == "" {
		_, oldUser.Password, err = secrets.SetMailUser(oldUser.Email(), password)
		if err != nil {
			return err
		}
		changed = true
	} else {
		hashed := oldSecret.Output
		if oldUser.Password != hashed {
			oldUser.Password = hashed
			changed = true
		}
	}

	if changed {
		if err := domainList.Save(); err != nil {
			return err
		}
		fmt.Printf("User '%s' was %s\n", oldUser.Email(), result)
	}

	return nil
}
