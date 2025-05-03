package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

type CommandWrapper struct {
	Cmd    *cli.Command
	Policy string // "prod", "dev", "any"
}

var (
	version   string
	buildTime string

	commandSet []CommandWrapper
)

var mainFlagCmds = cli.BoolFlag{
	Name:  "commands",
	Usage: T("main-flag-commands"),
}

var mainFlagDryRun = cli.BoolFlag{
	Name:  "dry",
	Usage: T("main-flag-dry-run"),
}

func main() {
	app := &cli.App{
		Name:    "gd-tools",
		Version: fmt.Sprintf("%s (built %s)", version, buildTime),
		Usage:   T("usage-main-app"),
		Flags:   []cli.Flag{&mainFlagCmds},
		Before: func(c *cli.Context) error {
			if c.Bool("commands") {
				for i, wrapper := range commandSet {
					if i > 0 {
						fmt.Print(" ")
					}
					fmt.Print(wrapper.Cmd.Name)
				}
				fmt.Println()
				os.Exit(0)
			}

			return nil
		},
		Commands: visibleCommands(),
		Action: func(c *cli.Context) error {
			fmt.Println(T("app-action-commands"))
			for _, wrapper := range commandSet {
				if CheckEnv(wrapper.Policy) {
					fmt.Printf("  %-20s %s\n", wrapper.Cmd.Name, wrapper.Cmd.Usage)
				}
			}
			return nil
		},
	}

	// Symlink-Unterstützung: z. B. gd-up → up
	if execName := filepath.Base(os.Args[0]); len(execName) > 3 && execName[:3] == "gd-" && execName != "gd-tools" {
		cmd := execName[3:]
		if len(os.Args) == 1 {
			os.Args = append(os.Args, cmd)
		}
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func AddSubCommand(cmd *cli.Command, policy string) {
	commandSet = append(commandSet, CommandWrapper{Cmd: cmd, Policy: policy})
	sort.Slice(commandSet, func(i, j int) bool {
		return commandSet[i].Cmd.Name < commandSet[j].Cmd.Name
	})
}

func visibleCommands() []*cli.Command {
	var cmds []*cli.Command
	for _, wrapper := range commandSet {
		if CheckEnv(wrapper.Policy) {
			cmds = append(cmds, wrapper.Cmd)
		}
	}
	return cmds
}

func ReadEnv() string {
	data, err := os.ReadFile("/etc/gd-tools-env")
	if err != nil {
		return "unknown"
	}

	return strings.TrimSpace(string(data))
}

func CheckEnv(wanted string) bool {
	if wanted == "any" {
		return true
	}

	return ReadEnv() == wanted
}
