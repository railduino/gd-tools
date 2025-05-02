package main

import (
	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandDeploy, "devOnly")
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

	deployScript, err := FileDeployRead()
	if err != nil {
		return err
	}

	var cmds []string
	for _, line := range deployScript.Commands {
		cmds = append(cmds, line)
	}

	return ShellCmds(dryRun, cmds)
}
