package bash

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/templates"
	"github.com/urfave/cli/v2"
)

const (
	SaveDir = "/etc/bash_completion.d"
)

var Describe = `The bash command generates a Bash completion script for gdt.

The generated script enables tab completion for commands, subcommands, and
flags, improving interactive usability on the development workstation.

The script can either be written to standard output or installed system-wide
into the Bash completion directory.

This command is intended to be run on the development system only.`

var Command = &cli.Command{
	Name:        "bash",
	Usage:       "Generate a Bash completion script for gdt",
	Description: Describe,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "save",
			Usage: "write to file instead of stdout",
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	path, err := os.Executable()
	if err != nil {
		return err
	}
	name := filepath.Base(path)

	data := struct {
		Name string
	}{
		Name: name,
	}

	bashPath := filepath.Join("assets", "completion.bash")
	out, err := templates.Parse(bashPath, false, data)
	if err != nil {
		return err
	}
	completionName := name + "_completion"

	if c.Bool("save") {
		path := filepath.Join(SaveDir, completionName)
		if err := os.WriteFile(path, out, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", completionName, err)
		}

		fmt.Printf("Completion script saved to: %s\n", path)
		return nil
	}

	_, err = os.Stdout.Write(out)
	return err
}
