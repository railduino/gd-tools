package account

import (
	"fmt"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/email"
	"github.com/urfave/cli/v2"
)

var ForwardCommand = &cli.Command{
	Name:  "forward",
	Usage: "add or delete forwarding targets for a mailbox",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.BoolFlag{
			Name:  "dismiss",
			Usage: "forward only (do not keep local copy)",
		},
	},
	ArgsUsage: "add <mailbox> <target>...\n   gdt account forward delete <mailbox> [<target>...]",
	Action:    ForwardRun,
}

func ForwardRun(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	if c.NArg() < 2 {
		return fmt.Errorf("[forward] missing action and/or mailbox")
	}

	action := c.Args().Get(0)
	mailbox := c.Args().Get(1)

	mb, err := email.MakeUser(mailbox)
	if err != nil {
		return fmt.Errorf("[forward] invalid mailbox '%s': %w", mailbox, err)
	}

	domainList, domainMap, err := email.GetDomains(nil)
	if err != nil {
		return err
	}

	domain, ok := domainMap[mb.Domain]
	if !ok {
		return fmt.Errorf("[forward] domain '%s' not found", mb.Domain)
	}

	key := mb.Email() // full email, because UserMap is keyed by Email()
	u, ok := domain.UserMap[key]
	if !ok || u == nil {
		return fmt.Errorf("[forward] mailbox '%s' not found", key)
	}

	switch action {
	case "add":
		if c.NArg() < 3 {
			return fmt.Errorf("[forward] add requires at least one target address")
		}

		// Explicitly set dismiss when adding forwards
		u.Dismiss = c.Bool("dismiss")

		for i := 2; i < c.NArg(); i++ {
			if err := u.AddForward(c.Args().Get(i)); err != nil {
				return fmt.Errorf("[forward] %s: %w", u.Email(), err)
			}
		}

	case "delete":
		// delete <mailbox> => clear all forwards
		if c.NArg() == 2 {
			u.ClearForwards()
		} else {
			for i := 2; i < c.NArg(); i++ {
				if err := u.DeleteForward(c.Args().Get(i)); err != nil {
					return fmt.Errorf("[forward] %s: %w", u.Email(), err)
				}
			}
		}

	default:
		return fmt.Errorf("[forward] invalid action '%s'", action)
	}

	if err := domainList.Save(); err != nil {
		return err
	}

	return nil
}
