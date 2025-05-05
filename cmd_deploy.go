package main

import (
	"fmt"
	"os"
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

	configCopy := fmt.Sprintf("%s --chmod=400 %s %s:/etc", rsyncRoot, SystemConfigName, rootUser)
	if err := ShellCmd(dryRun, configCopy); err != nil {
		return err
	}

	toolsUser := fmt.Sprintf("gd-tools@%s", hostName)
	rsyncFlagList := []string{
		"--chown=gd-tools:gd-tools",
		"--exclude=letsencrypt",
		"--exclude=secrets.json",
		"--exclude=data",
		"--exclude=" + SystemConfigName,
	}
	rsyncFlags := strings.Join(rsyncFlagList, " ")
	projectCopy := fmt.Sprintf("rsync -avz %s %s/ %s:projects", rsyncFlags, localPath, toolsUser)
	if err := ShellCmd(dryRun, projectCopy); err != nil {
		return err
	}

	projects, err := ProjectLoadAll()
	if err != nil {
		return err
	}
	for _, p := range projects {
		localDataPath := filepath.Join(p.GetName(), "data")
		if stat, err := os.Stat(localDataPath); err == nil && stat.IsDir() {
			remoteDataPath := fmt.Sprintf("%s:/var/gd-tools/data/%s", toolsUser, p.GetName())
			rsyncCmd := fmt.Sprintf("rsync -avz --update --chown=gd-tools:gd-tools %s/ %s",
				localDataPath, remoteDataPath)
			if err := ShellCmd(dryRun, rsyncCmd); err != nil {
				return err
			}
		}
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
