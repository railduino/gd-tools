package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/net/publicsuffix"
	"gopkg.in/ini.v1"
)

func init() {
	AddSubCommand(commandSetup, "devOnly")
}

var defaultPackages = []string{
	"openssh-server",
	"xauth",
	"git",
	"rsync",
	"file",
	"tree",
	"unzip",
	"bzip2",
	"curl",
	"wget",
	"vim",
	"jq",
	"plocate",
	"apt-transport-https",
	"software-properties-common",
	"ca-certificates",
	"gnupg2",
	"lsb-release",
	"dirmngr",
	"docker-ce",
	"docker-ce-cli",
	"containerd.io",
	"docker-buildx-plugin",
	"docker-compose-plugin",
}

var defaultMounts = []Mount{
	{
		Provider:   "Hetzner",
		Identifier: "1234567890",
		Mountpoint: "/var/gd-tools",
	},
	{
		Provider:   "RAID",
		Identifier: "UUID-1234-5678-ABCD",
		Mountpoint: "/var/gd-tools",
	},
}

var commandSetup = &cli.Command{
	Name:        "setup",
	Usage:       T("setup-cmd-usage"),
	Description: T("setup-cmd-describe"),
	Flags:       []cli.Flag{},
	ArgsUsage:   "<hostname>",
	Action:      runSetup,
}

func runSetup(c *cli.Context) error {
	if c.NArg() < 1 {
		msg := T("setup-err-missing-host")
		return fmt.Errorf(msg)
	}
	dstPath := c.Args().First()

	// By convention the directory is the hostname
	hostName := filepath.Base(dstPath)
	domainName, err := publicsuffix.EffectiveTLDPlusOne(hostName)
	if err != nil {
		return err
	}

	// try to get the sysadmin from .gitconfig
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	gitConfigPath := filepath.Join(homeDir, ".gitconfig")
	cfg, err := ini.Load(gitConfigPath)
	if err != nil {
		return err
	}
	sysAdmin := cfg.Section("user").Key("email").String()

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
	timezone, err := os.ReadFile("/etc/timezone")
	if err != nil {
		return err
	}
	systemConfig := SystemConfig{
		Version:    version,
		TimeZone:   strings.TrimSpace(string(timezone)),
		SwapSpace:  0,
		HostName:   hostName,
		DomainName: domainName,
		SshPort:    "OpenSSH",
		SysAdmin:   sysAdmin,
		Packages:   defaultPackages,
		Mounts:     defaultMounts,
	}
	if err := systemConfig.Save(); err != nil {
		return err
	}

	fmt.Println(T("setup-step-deploy"))
	myself, err := os.Executable()
	if err != nil {
		return err
	}
	myself, err = filepath.EvalSymlinks(myself)
	if err != nil {
		return err
	}

	sync_user := fmt.Sprintf("root@%s", hostName)
	sync_excl := "--exclude=logs --exclude=secrets.json --exclude=deploy.json"
	sync_cmds := []string{
		fmt.Sprintf("rsync -avz %s %s/ %s:/etc/gd-tools", sync_excl, dstPath, sync_user),
		fmt.Sprintf("rsync -avz %s/ %s:/usr/local/bin", myself, sync_user),
		fmt.Sprintf("ssh %s /usr/local/bin/gd-tools update", sync_user),
	}

	deployScript := DeployScript{
		Commands: sync_cmds,
	}
	if err := deployScript.Save(); err != nil {
		return err
	}

	return nil
}
