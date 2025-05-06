package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

type ServeData struct {
	AdminUser string `json:"admin_user"`
	AdminMail string `json:"admin_mail"`
	AdminPswd string `json:"admin_pswd"`

	ServePort string `json:"serve_port"`
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
		msg := T("serve-avoid-root")
		return fmt.Errorf(msg)
	}

	return nil
}
