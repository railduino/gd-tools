package main

import (
	"fmt"
	"os"
	"path/filepath"

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
	toolsUser := fmt.Sprintf("gd-tools@%s", hostName)

	// Deploy binary
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}
	if err := DeployLocal(dryRun, execPath, "/usr/local/bin", rootUser, "755"); err != nil {
		return err
	}

	// Deploy system config
	if err := DeployLocal(dryRun, SystemConfigName, "/etc", rootUser, "400"); err != nil {
		return err
	}

	// Deploy static templates
	if err := DeployTemplate(dryRun, "ssl-params.conf", "/etc/nginx/snippets/ssl-params.conf", rootUser, "444"); err != nil {
		return err
	}
	if err := DeployTemplate(dryRun, "nginx.conf", "/etc/nginx/nginx.conf", rootUser, "444"); err != nil {
		return err
	}

	// Deploy project tree (excluding sensitive data)
	rsyncProjects := DeployRsync{
		DryRun: dryRun,
		Flags: []string{
			"--chown=gd-tools:gd-tools",
			"--exclude=letsencrypt",
			"--exclude=secrets.json",
			"--exclude=data",
			"--exclude=" + SystemConfigName,
		},
		Local:    localPath + "/",
		Receiver: toolsUser,
		Remote:   "projects",
	}
	if err := rsyncProjects.Execute(); err != nil {
		return err
	}

	// Deploy project-specific data dirs
	projects, err := ProjectLoadAll()
	if err != nil {
		return err
	}
	for _, p := range projects {
		dataPath := filepath.Join(p.GetName(), "data")
		if stat, err := os.Stat(dataPath); err == nil && stat.IsDir() {
			rsyncData := DeployRsync{
				DryRun:   dryRun,
				Flags:    []string{"--chown=gd-tools:gd-tools", "--update"},
				Local:    dataPath + "/",
				Receiver: toolsUser,
				Remote:   "/var/gd-tools/data/" + p.GetName(),
			}
			if err := rsyncData.Execute(); err != nil {
				return err
			}
		}
	}

	// Fetch certs from target before overwrite
	DeployFetchLetsEncrypt(dryRun, rootUser)

	// Push certs if available locally
	if stat, err := os.Stat("letsencrypt"); err == nil && stat.IsDir() {
		rsyncCerts := DeployRsync{
			DryRun:   dryRun,
			Flags:    []string{"--chown=root:root"},
			Local:    "letsencrypt/",
			Receiver: rootUser,
			Remote:   "/etc/letsencrypt",
		}
		if err := rsyncCerts.Execute(); err != nil {
			return err
		}
	}

	return nil
}
