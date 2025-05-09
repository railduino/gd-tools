package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/ini.v1"
)

func init() {
	AddSubCommand(commandSetup, "dev")
}

var setupFlagHetzner = cli.StringFlag{
	Name:  "hetzner",
	Usage: T("setup-flag-hetzner"),
}

var setupFlagRAID = cli.StringFlag{
	Name:  "raid",
	Usage: T("setup-flag-raid"),
}

var commandSetup = &cli.Command{
	Name:        "setup",
	Usage:       T("setup-cmd-usage"),
	Description: T("setup-cmd-describe"),
	Flags: []cli.Flag{
		&setupFlagHetzner,
		&setupFlagRAID,
	},
	ArgsUsage: "<hostname>",
	Action:    runSetup,
}

func runSetup(c *cli.Context) error {
	if c.NArg() < 1 {
		msg := T("setup-err-missing-host")
		return fmt.Errorf(msg)
	}
	dstPath := c.Args().First()

	systemFile := filepath.Join(dstPath, SystemConfigName)
	if _, err := os.Stat(systemFile); err == nil {
		msg := Tf("setup-err-host-exist", dstPath)
		return fmt.Errorf(msg)
	}

	// by convention the directory is the hostname
	hostName := filepath.Base(dstPath)
	domainName, err := publicsuffix.EffectiveTLDPlusOne(hostName)
	if err != nil {
		return err
	}

	// collect default DEB packages
	packages, err := TemplateLines("packages.txt", "#")
	if err != nil {
		return err
	}

	// check for filesystems to be mounted
	// N.B. mounts given here are mutually exclusive
	var mounts []Mount
	if volume := c.String("hetzner"); volume != "" {
		mount := Mount{
			Provider:   "Hetzner",
			Identifier: volume,
			Mountpoint: SystemVarMount,
		}
		mounts = append(mounts, mount)
	} else if device := c.String("raid"); device != "" {
		mount := Mount{
			Provider:   "RAID",
			Identifier: device,
			Mountpoint: SystemVarMount,
		}
		mounts = append(mounts, mount)
	}

	// get the sysadmin email from .gitconfig if possible
	sysAdmin := fmt.Sprintf("admin@%s", domainName)
	homeDir, err := os.UserHomeDir()
	if err == nil {
		gitConfigPath := filepath.Join(homeDir, ".gitconfig")
		cfg, err := ini.Load(gitConfigPath)
		if err == nil {
			sysAdmin = cfg.Section("user").Key("email").String()
		}
	}

	fmt.Println(T("setup-step-mkdir"))
	if err := os.MkdirAll(dstPath, 0755); err != nil {
		return err
	}
	if err := os.Chdir(dstPath); err != nil {
		return err
	}

	fmt.Println(T("setup-step-system"))
	timeZone, err := FileGetLine("/etc/timezone")
	if err != nil {
		return err
	}

	systemConfig := SystemConfig{
		Version:    version,
		TimeZone:   timeZone,
		HostName:   hostName,
		DomainName: domainName,
		SwapSpace:  0,
		SysAdmin:   sysAdmin,
		Packages:   packages,
		Mounts:     mounts,
	}

	if err := systemConfig.Save(); err != nil {
		return err
	}

	return nil
}
