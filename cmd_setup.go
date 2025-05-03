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

	if _, err := os.Stat(dstPath); err == nil {
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
	mountpoint, err := GetDataRoot("prod", "")
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
	} else if device := c.String("raid"); device != "" {
		mount := Mount{
			Provider:   "RAID",
			Identifier: device,
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
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	var deployCmds []string
	projectRoot, _ := GetProjectRoot("prod")
	toolUser := fmt.Sprintf("gd-tools@%s", hostName)
	syncExcl := "--exclude=logs --exclude=secrets.json --exclude=deploy.json"
	syncProg := "--chown=gd-tools:gd-tools --chmod=0755"
	userCmds := []string{
		fmt.Sprintf("rsync -avz %s %s/ %s:%s", syncExcl, dstPath, toolUser, projectRoot),
		fmt.Sprintf("rsync -avz %s %s %s:/usr/local/bin", syncProg, execPath, toolUser),
	}
	for _, cmd := range userCmds {
		deployCmds = append(deployCmds, cmd)
	}

	deployScript := DeployScript{
		Commands: deployCmds,
	}
	if err := deployScript.Save(); err != nil {
		return err
	}

	return nil
}
