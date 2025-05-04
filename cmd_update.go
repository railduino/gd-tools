package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	AddSubCommand(commandUpdate, "prod")
}

var commandUpdate = &cli.Command{
	Name:        "update",
	Usage:       T("update-cmd-usage"),
	Description: T("update-cmd-describe"),
	Action:      runUpdate,
}

func runUpdate(c *cli.Context) error {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		return err
	}
	fmt.Println("projectRoot: ", projectRoot)

	entries, err := os.ReadDir(projectRoot)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fmt.Println("entry: ", entry.Name())
		if !entry.IsDir() {
			continue
		}
		projectPath := filepath.Join(projectRoot, entry.Name())
		if _, err := os.Stat(filepath.Join(projectPath, "config.json")); err != nil {
			continue
		}

		fmt.Println("updating ...")
		if err := updateProject(projectPath); err != nil {
			fmt.Fprintf(os.Stderr, "Fehler bei %s: %v\n", entry.Name(), err)
		}
	}

	return nil
}

func updateProject(projectPath string) error {
	uid := os.Geteuid()
	gid := os.Getegid()

	fmt.Println("projectPath: ", projectPath)
	env := fmt.Sprintf("GDTOOLS_UID=%d\nGDTOOLS_GID=%d\n", uid, gid)
	envPath := filepath.Join(projectPath, ".env")

	return os.WriteFile(envPath, []byte(env), 0644)
}
