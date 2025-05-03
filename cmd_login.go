package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandLogin, "dev")
}

var loginFlagRoot = cli.BoolFlag{
	Name:  "root",
	Usage: T("login-flag-root"),
}

var commandLogin = &cli.Command{
	Name:        "login",
	Usage:       T("login-cmd-usage"),
	Description: T("login-cmd-describe"),
	Flags: []cli.Flag{
		&loginFlagRoot,
	},
	Action: runLoginServer,
}

func runLoginServer(c *cli.Context) error {
	localPath, err := os.Getwd()
	if err != nil {
		return err
	}
	hostName := filepath.Base(localPath)

	userName := "gd-tools"
	if c.Bool("root") {
		userName = "root"
	}

	cmd := exec.Command("ssh", fmt.Sprintf("%s@%s", userName, hostName))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
