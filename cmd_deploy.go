package main

import (
	_ "bytes"
	"fmt"
	"os"
	_ "os/exec"
	"path/filepath"
	_ "strings"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandDeploy, "dev")
}

var commandDeploy = &cli.Command{
	Name:        "deploy",
	Usage:       T("deploy-cmd-usage"),
	Description: T("deploy-cmd-describe"),
	Flags: []cli.Flag{
		&mainFlagDryRun,
	},
	Action: runDeploy,
}

func runDeploy(c *cli.Context) error {
	dryRun := c.Bool("dry")

	localPath, err := os.Getwd()
	if err != nil {
		return err
	}
	hostName := filepath.Base(localPath)
	rootUser := fmt.Sprintf("root@%s", hostName)
	rsyncRoot := "rsync -avz --chown=root:root"

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}
	execCopy := fmt.Sprintf("%s --chmod=755 %s %s:/usr/local/bin", rsyncRoot, execPath, rootUser)
	if err := ShellCmd(dryRun, execCopy); err != nil {
		return err
	}

	deployFetchLetsEncrypt(dryRun, rootUser)

	configCopy := fmt.Sprintf("%s --chmod=400 %s %s:/etc", rsyncRoot, SystemConfigFile, rootUser)
	if err := ShellCmd(dryRun, configCopy); err != nil {
		return err
	}

	toolUser := fmt.Sprintf("gd-tools@%s", hostName)
	rsyncUser := "rsync -avz --chown=gd-tools:gd-tools"
	rsyncExcl := "--exclude=letsencrypt --exclude=secrets.json --exclude=" + SystemConfigFile
	projectCopy := fmt.Sprintf("%s %s %s/ %s:projects", rsyncUser, rsyncExcl, localPath, toolUser)
	if err := ShellCmd(dryRun, projectCopy); err != nil {
		return err
	}

	if _, err := os.Stat("letsencrypt"); err == nil {
		certCopy := fmt.Sprintf("rsync -avz --chown=root:root letsencrypt/ %s:/etc/letsencrypt", rootUser)
		if err := ShellCmd(dryRun, certCopy); err != nil {
			return err
		}
	}

	return nil
}

func deployFetchLetsEncrypt(dryRun bool, rootUser string) {
	rsyncCmd := fmt.Sprintf("rsync -avz %s:/etc/letsencrypt/ letsencrypt", rootUser)

	if err := ShellCmd(dryRun, rsyncCmd); err != nil {
		fmt.Println("Ignore error:", err)
	}
}
