package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandDelete, "dev")
}

var deleteFlagForce = cli.BoolFlag{
	Name:    "force",
	Aliases: []string{"f"},
	Usage:   T("delete-flag-force"),
}

var commandDelete = &cli.Command{
	Name:        "delete",
	Usage:       T("delete-cmd-usage"),
	Description: T("delete-cmd-describe"),
	ArgsUsage:   "<project>",
	Flags: []cli.Flag{
		&deleteFlagForce,
		&mainFlagDryRun,
	},
	Action: runDeleteProject,
}

func runDeleteProject(c *cli.Context) error {
	if _, err := ReadSystemConfig(true); err != nil {
		return err
	}

	dryRun := c.Bool("dry")
	force := c.Bool("force")

	args := c.Args().Slice()
	if len(args) != 1 {
		msg := T("delete-err-no-arg")
		return fmt.Errorf(msg)
	}
	projectPath := args[0]

	info, err := os.Stat(projectPath)
	if err != nil {
		if os.IsNotExist(err) {
			msg := Tf("delete-err-not-found", projectPath)
			return fmt.Errorf(msg)
		}
		return err
	}
	if !info.IsDir() {
		msg := Tf("delete-err-not-dir", projectPath)
		return fmt.Errorf(msg)
	}

	if !force {
		msg := T("delete-err-no-force")
		return fmt.Errorf(msg)
	}

	cmdStr := fmt.Sprintf("rm -rf %s", projectPath)
	return ShellCmd(dryRun, cmdStr)
}
