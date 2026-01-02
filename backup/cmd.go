package backup

import (
	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/utils"
	"github.com/urfave/cli/v2"
)

var Describe = `The backup command prepares a reliable backup and restore strategy.

BorgBackup is a Deduplicating archiver with compression and encryption (https://www.borgbackup.org/).

A deduplicated snapshot of /var/gd-tools/data is created nightly via borg exec.
Tools like bb.list, bb.mount, and others are provided to manage the Borg repository.
it is highly recommended to keep the backup separate from the server.
This might be a Hetzner Storage Box, or just w Raspberry Pi behind a Fritz!Box next door.
It is the responsibility of the administrator to ensure proper SSH access,
including both known_hosts and authorized_keys. Check before relying on it!

🔒 Reminder: The only backup you ever need is the one you did not make.`

var Command = &cli.Command{
	Name:        "backup",
	Usage:       "Install reliable backup and restore with BorgBackup",
	Description: Describe,
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		&cli.StringFlag{
			Name:  "passphrase",
			Usage: "passphrase for Borg (uses repokey-blake2)",
		},
		&cli.StringFlag{
			Name:  "remote-path",
			Usage: "remote mirror, reachable with rsync over SSH",
		},
		&cli.StringFlag{
			Name:  "remote-shell",
			Usage: "rsync shell for '-e' flag to access remote",
		},
		&cli.IntFlag{
			Name:  "days",
			Usage: "how many days to keep before purging",
			Value: 7,
		},
		&cli.IntFlag{
			Name:  "weeks",
			Usage: "how many weeks to keep before purging",
			Value: 4,
		},
		&cli.IntFlag{
			Name:  "months",
			Usage: "how many months to keep before purging",
			Value: 12,
		},
		&cli.IntFlag{
			Name:  "hour",
			Usage: "backup and mirror hour, usually at night",
			Value: 3,
		},
		&cli.IntFlag{
			Name:  "minute",
			Usage: "backup and mirror minute, usually at night",
			Value: 30,
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	_, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

	passphrase := c.String("passphrase")
	if passphrase == "" {
		passphrase, err = utils.CreatePassword(24)
		if err != nil {
			return err
		}
	}

	entry := config.Backup{
		PassPhrase:  passphrase,
		RemotePath:  c.String("remote-path"),
		RemoteShell: c.String("remote-shell"),
		KeepDays:    Validate(c.Int("days"), 1, 365),
		KeepWeeks:   Validate(c.Int("weeks"), 1, 52),
		KeepMonths:  Validate(c.Int("months"), 1, 36),
		CronHour:    Validate(c.Int("hour"), 0, 23),
		CronMinute:  Validate(c.Int("minute"), 0, 59),
	}

	if err := entry.Save(); err != nil {
		return err
	}

	return nil
}

func Validate(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
