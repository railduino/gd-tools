package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandLogin, "devOnly")
}

var commandLogin = &cli.Command{
	Name:        "login",
	Usage:       T("login-cmd-usage"),
	Description: T("login-cmd-describe"),
	Action:      runLoginServer,
}

func runLoginServer(c *cli.Context) error {
	localPath, err := os.Getwd()
	if err != nil {
		return err
	}

	scriptFile := filepath.Join(localPath, "deploy.json")
	if _, err := os.Stat(scriptFile); err != nil {
		msg := T("err-missing-json")
		return fmt.Errorf(msg)
	}

	fmt.Println("Login to:", filepath.Base(localPath))

	return nil
}
