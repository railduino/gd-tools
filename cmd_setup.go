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
	AddSubCommand(commandSetup, "devOnly")
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

	// by convention the directory is the hostname
	hostName := filepath.Base(dstPath)
	domainName, err := publicsuffix.EffectiveTLDPlusOne(hostName)
	if err != nil {
		return err
	}

	// collect default DEB packages
	packages, err := TemplateLines("packages.txt")
	if err != nil {
		return err
	}

	// check for mounted filesystems
	// N.B. mounts given here are mutually exclusive
	var mounts []Mount
	mountpoint, err := GetDataRoot(true, "")
	if err != nil {
		return err
	}
	if volume := c.String("hetzner"); volume != "" {
		mount := Mount{
			Provider:   "Hetzner",
			Identifier: volume,
			Mountpoint: mountpoint,
		}
		mounts = append(mounts, mount)
	} else if uuid := c.String("raid"); uuid != "" {
		mount := Mount{
			Provider:   "RAID",
			Identifier: uuid,
			Mountpoint: mountpoint,
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

	if _, err := os.Stat(dstPath); err == nil {
		msg := Tf("setup-err-host-exist", dstPath)
		return fmt.Errorf(msg)
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
		SwapSpace:  0,
		HostName:   hostName,
		DomainName: domainName,
		SshPort:    "OpenSSH",
		SysAdmin:   sysAdmin,
		Packages:   packages,
		Mounts:     mounts,
	}
	if err := systemConfig.Save(); err != nil {
		return err
	}

	fmt.Println(T("setup-step-deploy"))
	myExec, err := os.Executable()
	if err != nil {
		return err
	}
	myExec, err = filepath.EvalSymlinks(myExec)
	if err != nil {
		return err
	}

	syncRoot, err := GetProjectRoot(true)
	if err != nil {
		return err
	}
	syncUser := fmt.Sprintf("root@%s", hostName)
	syncExcl := "--exclude=logs --exclude=secrets.json --exclude=deploy.json"
	syncCmds := []string{
		fmt.Sprintf("rsync -avz %s %s/ %s:%s", syncExcl, dstPath, syncUser, syncRoot),
		fmt.Sprintf("rsync -avz %s/ %s:/usr/local/bin", myExec, syncUser),
		fmt.Sprintf("ssh %s /usr/local/bin/gd-tools update", syncUser),
	}

	deployScript := DeployScript{
		Commands: syncCmds,
	}
	if err := deployScript.Save(); err != nil {
		return err
	}

	return nil
}
