package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/railduino/gd-tools/agent"
	"github.com/urfave/cli/v2"
)

var (
	ProgName string
	InstName string
)

func main() {
	ProgName = filepath.Base(os.Args[0])
	name, ok := strings.CutPrefix(ProgName, "occ-")
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: program name must start with 'occ-'\n")
		os.Exit(1)
	}
	InstName = name

	app := &cli.App{
		Name:  ProgName,
		Usage: "Run Nextcloud OCC commands for project '" + InstName + "'",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "fix",
				Usage: "fix various tables and indices",
			},
			&cli.BoolFlag{
				Name:  "check",
				Usage: "Run update:check",
			},
			&cli.BoolFlag{
				Name:  "update",
				Usage: "Run app:update --all",
			},
			&cli.BoolFlag{
				Name:  "upgrade",
				Usage: "Run updater.phar with apc.enable_cli=1",
			},
		},
		EnableBashCompletion: true,
		Action:               Run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func Run(c *cli.Context) error {
	nc := &agent.Nextcloud{
		Name: InstName,
	}

	var resp agent.Response
	var value string
	var err error
	switch {
	case c.Bool("check"):
		_, err = nc.RunOCC(&resp, "update:check")
	case c.Bool("update"):
		_, err = nc.RunOCC(&resp, "app:update", "--all")
	default:
		args := c.Args().Slice()
		if len(args) == 0 {
			_, err = nc.RunOCC(&resp, "")
		} else {
			value, err = nc.RunOCC(&resp, args[0], args[1:]...)
			if value != "" {
				fmt.Println(value)
			}
		}
	}
	if err != nil {
		return err
	}
	for _, line := range resp.Result {
		fmt.Println(line)
	}

	return nil
}
