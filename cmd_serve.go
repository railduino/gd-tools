package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	ServeConfigName = "gd-tools-serve.conf"
)

type ServeConfig struct {
	AdminUser string `json:"admin_user"`
	AdminMail string `json:"admin_mail"`
	AdminPswd string `json:"admin_pswd"`

	ServePort string `json:"serve_port"`

	ImprintURL string `json:"imprint_url"`
	ProtectURL string `json:"protect_url"`
}

func init() {
	AddSubCommand(commandServe, "prod")
}

var commandServe = &cli.Command{
	Name:        "serve",
	Usage:       T("serve-cmd-usage"),
	Description: T("serve-cmd-describe"),
	Flags:       []cli.Flag{},
	Action:      runServe,
}

func runServe(c *cli.Context) error {
	if euid := os.Geteuid(); euid == 0 {
		msg := T("serve-not-as-root")
		return fmt.Errorf(msg)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.Chdir(homeDir); err != nil {
		return err
	}

	if err := InitServeLocale(); err != nil {
		return err
	}

	if err := InitServeDB(); err != nil {
		return err
	}

	if err := InitServeWeb("127.0.0.1:3000"); err != nil {
		return err
	}

	return nil
}
