package main

import (
	"fmt"
	"os"

	"github.com/railduino/gd-tools/account"
	"github.com/railduino/gd-tools/backup"
	"github.com/railduino/gd-tools/bash"
	"github.com/railduino/gd-tools/basics"
	"github.com/railduino/gd-tools/cert"
	"github.com/railduino/gd-tools/deploy"
	"github.com/railduino/gd-tools/hello"
	"github.com/railduino/gd-tools/install"
	"github.com/railduino/gd-tools/login"
	"github.com/railduino/gd-tools/nextcloud"
	"github.com/railduino/gd-tools/ocis"
	"github.com/railduino/gd-tools/redirect"
	"github.com/railduino/gd-tools/setup"
	"github.com/railduino/gd-tools/ssh"
	"github.com/railduino/gd-tools/status"
	"github.com/railduino/gd-tools/sync"
	"github.com/railduino/gd-tools/update"
	"github.com/railduino/gd-tools/wbce"
	"github.com/railduino/gd-tools/wordpress"
	"github.com/urfave/cli/v2"
)

var (
	version string // will be loaded via -ldflags
)

func main() {
	app := &cli.App{
		Name:     "gdt",
		Version:  fmt.Sprintf("%s", version),
		Usage:    "gdt (short for gd-tools or go-deployment-tools) helps managing Ubuntu/Debian servers",
		Commands: getCommands(),
		Action: func(c *cli.Context) error {
			return cli.ShowAppHelp(c)
		},
		EnableBashCompletion: true,
		BashComplete: func(c *cli.Context) {
			fmt.Println("account")
			fmt.Println("backup")
			fmt.Println("bash")
			fmt.Println("basics")
			fmt.Println("cert")
			fmt.Println("deploy")
			fmt.Println("hello")
			fmt.Println("install")
			fmt.Println("login")
			fmt.Println("nextcloud")
			fmt.Println("ocis")
			fmt.Println("redirect")
			fmt.Println("setup")
			fmt.Println("ssh")
			fmt.Println("status")
			fmt.Println("sync")
			fmt.Println("update")
			fmt.Println("wbce")
			fmt.Println("wordpress")
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func getCommands() []*cli.Command {
	return []*cli.Command{
		account.Command,
		backup.Command,
		bash.Command,
		basics.Command,
		cert.Command,
		deploy.Command,
		hello.Command,
		install.Command,
		login.Command,
		nextcloud.Command,
		ocis.Command,
		redirect.Command,
		setup.Command,
		ssh.Command,
		status.Command,
		sync.Command,
		update.Command,
		wbce.Command,
		wordpress.Command,
	}
}
