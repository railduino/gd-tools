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
	name, ok := strings.CutPrefix(ProgName, "wp-")
	if !ok {
		fmt.Fprintf(os.Stderr, "Error: program name must start with 'wp-'\n")
		os.Exit(1)
	}
	InstName = name

	app := &cli.App{
		Name:  ProgName,
		Usage: "Run WP-CLI commands for project '" + InstName + "'",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "dir-name",
				Usage:   "subdirectory of the WordPress instance",
				Value:   "wordpress",
				EnvVars: []string{"GD_WP_DIRNAME"},
			},
			&cli.BoolFlag{
				Name:  "cron",
				Usage: "Run WordPress cron (due events)",
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
	wp := &agent.WordPress{
		Instance: InstName,
		DirName:  c.String("dir-name"),
	}

	var resp agent.Response
	var value string
	var err error

	switch {
	case c.Bool("cron"):
		_, err = wp.RunWPCLI(&resp, "cron", "event", "run", "--due-now")
	default:
		args := c.Args().Slice()
		if len(args) == 0 {
			fmt.Fprintf(os.Stderr, "No command given\n")
			return nil
		}
		value, err = wp.RunWPCLI(&resp, args[0], args[1:]...)
		if value != "" {
			fmt.Println(value)
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
