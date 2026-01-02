package bash

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/railduino/gd-tools/config"
	"github.com/railduino/gd-tools/templates"
	"github.com/urfave/cli/v2"
)

const (
	SaveDir = "/etc/bash_completion.d"
)

var Command = &cli.Command{
	Name:  "bash",
	Usage: "Generate a Bash completion script for gdt",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "save",
			Usage: "write to file instead of stdout",
		},
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, err := config.ReadConfig(c)
	if err != nil {
		return err
	}

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

	out, err := templates.Parse("completion.bash", cfg.Verbose, data)
	if err != nil {
		return err
	}
	completionName := name + "_completion"

	if c.Bool("save") {
		path := filepath.Join(SaveDir, completionName)
		if err := os.WriteFile(path, out, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", completionName, err)
		}

		cfg.Sayf("Completion script saved to: %s", path)
		return nil
	}

	_, err = os.Stdout.Write(out)
	return err
}
