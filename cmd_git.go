package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandGit, "devOnly")
}

var commandGit = &cli.Command{
	Name:        "git",
	Usage:       T("git-cmd-usage"),
	Description: T("git-cmd-describe"),
	ArgsUsage:   "<repo-url>",
	Flags:       []cli.Flag{&mainFlagDryRun},
	Action:      runGitInit,
}

func runGitInit(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("Fehler: fehlende Repository-URL")
	}

	repo_url := c.Args().First()
	dry_run := c.Bool("dry")

	// Prüfen ob bereits ein Git-Repository existiert
	if _, err := os.Stat(".git"); err == nil {
		return fmt.Errorf("Fehler: Hier existiert bereits ein Git-Repository (.git Verzeichnis)")
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("Fehler beim Prüfen von .git: %w", err)
	}

	cmds := []string{
		"git init",
		"git branch -M main",
		"git remote add origin " + repo_url,
		"git add .",
		"git commit -m Initial commit",
	}

	if err := ShellCmds(dry_run, cmds); err != nil {
		return err
	}

	fmt.Println("Git-Repository initialisiert und erster Commit erstellt.")
	return nil
}
