package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandLinks, "any")
}

var linksFlagDir = cli.StringFlag{
	Name:  "dir",
	Value: "/usr/local/bin",
	Usage: T("links-flag-dir"),
}

var commandLinks = &cli.Command{
	Name:        "links",
	Usage:       T("links-cmd-usage"),
	Description: T("links-cmd-describe"),
	Flags: []cli.Flag{
		&mainFlagDryRun,
		&linksFlagDir,
	},
	Action: runLinksCommand,
}

func runLinksCommand(c *cli.Context) error {
	bin_dir := c.String("dir")
	dry_run := c.Bool("dry")

	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	exeName := filepath.Base(exePath)

	if err := os.Chdir(bin_dir); err != nil {
		return err
	}

	validLinks := make(map[string]bool)
	for _, wrapper := range commandSet {
		if wrapper.Cmd.Hidden || wrapper.Cmd.Name == "help" {
			continue
		}
		linkName := "gd-" + wrapper.Cmd.Name
		validLinks[linkName] = true

		needCreate := true
		if fi, err := os.Lstat(linkName); err == nil && fi.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(linkName)
			if err == nil && target == exeName {
				needCreate = false
			}
		}

		if needCreate {
			link_cmd := fmt.Sprintf("ln -nfs %s %s", exeName, linkName)
			if err := ShellCmd(dry_run, link_cmd); err != nil {
				return err
			}
		}
	}

	entries, err := os.ReadDir(".")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasPrefix(name, "gd-") {
			continue
		}
		if _, ok := validLinks[name]; ok {
			continue
		}

		fi, err := entry.Info()
		if err != nil {
			return err
		}
		if fi.Mode()&os.ModeSymlink == 0 {
			continue
		}

		rm_cmd := fmt.Sprintf("rm %s", name)
		if err := ShellCmd(dry_run, rm_cmd); err != nil {
			return err
		}
	}

	return nil
}
