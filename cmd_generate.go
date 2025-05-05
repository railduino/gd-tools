package main

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/urfave/cli/v2"
)

type GenerateData struct {
	DataDir     string
	LogsDir     string
	ServicePort string
	ToolsUID    string
	DockerGID   string
}

// Registered project kinds (eg. wordpress, nextcloud, mail_host)
var generateKinds []*cli.Command

var generateCommand *cli.Command = commandGenerate

func init() {
	AddSubCommand(commandGenerate, "dev")
}

var generateDependsFlag = cli.StringSliceFlag{
	Name:    "depends",
	Aliases: []string{"d"},
	Usage:   T("generate-flag-depends"),
}

func RegisterProjectKind(cmd *cli.Command) {
	generateKinds = append(generateKinds, cmd)
	generateCommand.Subcommands = append(generateCommand.Subcommands, cmd)

	sort.Slice(generateKinds, func(i, j int) bool {
		return generateKinds[i].Name < generateKinds[j].Name
	})
}

var commandGenerate = &cli.Command{
	Name:        "generate",
	Usage:       T("generate-cmd-usage"),
	Description: T("generate-cmd-describe"),
	Aliases:     []string{"g"},
	Subcommands: []*cli.Command{},
	Action:      runGenerateDispatcher,
}

func runGenerateDispatcher(c *cli.Context) error {
	fmt.Println(T("generate-cmd-list-kinds"))

	for _, sub := range generateKinds {
		if sub.Name == "help" {
			continue
		}
		fmt.Printf("  %-20s %s\n", sub.Name, sub.Usage)
	}

	return nil
}

func GenerateCreateProject(c *cli.Context, kind string, unique bool) (*Project, error) {
	args := c.Args().Slice()
	if len(args) < 1 {
		msg := T("generate-err-prefix-missing")
		return nil, fmt.Errorf(msg)
	}
	prefixStr := args[0]

	number, err := GenerateParsePrefix(prefixStr)
	if err != nil {
		return nil, err
	}

	project := Project{
		Prefix:  fmt.Sprintf("%02d", number),
		Number:  number,
		PortStr: fmt.Sprintf("%d", number+8000),
		Kind:    kind,
	}

	if !unique {
		if len(args) < 2 {
			return nil, fmt.Errorf(T("generate-err-name-missing"))
		}
		project.Name = args[1]
	}

	if err := project.CheckConflict(unique); err != nil {
		return nil, err
	}

	return &project, nil
}

func GenerateParsePrefix(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf(T("generate-err-prefix-numeric"))
	}

	if n < 0 || n > 99 {
		return 0, fmt.Errorf(T("generate-err-prefix-numeric"))
	}

	return n, nil
}
