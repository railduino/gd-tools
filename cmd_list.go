package main

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandList, "any")
}

var listFlagStatus = cli.StringFlag{
	Name:    "status",
	Usage:   T("list-flag-status"),
	Aliases: []string{"s"},
}

var commandList = &cli.Command{
	Name:        "list",
	Usage:       T("list-cmd-usage"),
	Description: T("list-cmd-describe"),
	Flags: []cli.Flag{
		&listFlagStatus,
	},
	Action: runListProjects,
}

func runListProjects(c *cli.Context) error {
	projects, err := ProjectLoadAll(MainOnProd())
	if err != nil {
		return err
	}

	filterStatus := strings.ToLower(c.String("status"))

	fmt.Printf("%-20s %-15s %-20s %-10s\n", "PREFIX", "KIND", "NAME", "STATUS")
	fmt.Println(strings.Repeat("-", 70))

	for _, proj := range projects {
		status := "whatever" // TODO implement status "enabled,runing" etc.
		if status == "" {
			status = T("list-status-unknown")
		}

		if filterStatus != "" && !strings.Contains(status, filterStatus) {
			continue
		}

		fmt.Printf("%-20s %-15s %-20s %-10s\n", proj.Prefix, proj.Kind, proj.Name, status)
	}

	return nil
}
