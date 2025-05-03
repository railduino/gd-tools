package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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
	rootUser := fmt.Sprintf("root@%s", filepath.Base(localPath))

	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}
	execFlag := "-avz --chown=gd-tools:gd-tools --chmod=755"
	execCopy := fmt.Sprintf("rsync %s %s %s:/usr/local/bin", execFlag, execPath, rootUser)
	if err := ShellCmd(dryRun, execCopy); err != nil {
		return err
	}

	prepareName := "prepare-gd-tools.sh"
	projectRoot, err := GetProjectRoot("prod")
	if err != nil {
		return err
	}
	dataRoot, err := GetDataRoot("prod", "")
	if err != nil {
		return err
	}
	prepareData := struct {
		Dirs []string
	}{
		[]string{projectRoot, dataRoot},
	}
	prepareScript, err := TemplateParse(prepareName, prepareData)
	if err != nil {
		return err
	}

	if dryRun {
		prepareLines := strings.Split(string(prepareScript), "\n")
		for _, line := range prepareLines {
			fmt.Println(Tf("exec-dry-running", line))
		}
	} else {
		cmd := exec.Command("ssh", rootUser, "bash")
		cmd.Stdin = bytes.NewReader(prepareScript)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return err
		}
	}

	var cmds []string
	deployScript, err := FileDeployRead()
	if err != nil {
		return err
	}
	for _, line := range deployScript.Commands {
		cmds = append(cmds, line)
	}

	return ShellCmds(dryRun, cmds)
}
