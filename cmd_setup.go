package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
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
	hostname := filepath.Base(dstPath)

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
		Version:   version,
		TimeZone:  strings.TrimSpace(string(timezone)),
		SwapSpace: 0,
		HostName:  hostname,
		Packages:  defaultPackages,
		Mounts:    defaultMounts,
		SshPort:   "OpenSSH",
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

	sync_user := "root@" + hostname
	sync_excl := "--exclude=logs --exclude=deploy.json"
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
