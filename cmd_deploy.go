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

	configCopy := fmt.Sprintf("%s --chmod=400 %s %s:/etc", rsyncRoot, SystemConfigFile, rootUser)
	if err := ShellCmd(dryRun, configCopy); err != nil {
		return err
	}

	projectRoot, _ := GetProjectRoot("prod")
	toolUser := fmt.Sprintf("gd-tools@%s", hostName)
	rsyncUser := "rsync -avz --chown=gd-tools:gd-tools"
	rsyncExcl := "--exclude=logs --exclude=" + SystemConfigFile
	projectCopy := fmt.Sprintf("%s %s %s/ %s:%s", rsyncUser, rsyncExcl, localPath, toolUser, projectRoot)
	if err := ShellCmd(dryRun, projectCopy); err != nil {
		return err
	}

	return nil
}
