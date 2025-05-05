package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	RegisterProjectKind(commandGenerateMaintenance)
}

var commandGenerateMaintenance = &cli.Command{
	Name:        "maintenance",
	Usage:       T("install-maintenance-usage"),
	Description: T("install-maintenance-describe"),
	Flags:       []cli.Flag{},
	Action:      runGenerateMaintenance,
}

func runGenerateMaintenance(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf(T("generate-err-missing-prefix"))
	}

	project := Project{
		Prefix: args[0],
		Kind:   "maintenance",
		Name:   "",
	}
	if err := project.CheckConflict(true); err != nil {
		return err
	}

	fmt.Println(T("generate-create-dir"))
	projectPath, err := project.GetPath()
	if err != nil {
		return err
	}
	dataPath := filepath.Join(projectPath, "data")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return err
	}

	fmt.Println(T("generate-create-config"))
	project.Config = Config{
		IsEnabled: false,
		DependsOn: []string{},
		Networks:  []string{},
	}
	if err := project.SaveConfig(); err != nil {
		return err
	}

	fmt.Println(T("generate-create-index"))
	indexPath := filepath.Join("maintenance", "index.html")
	indexData, err := TemplateParse(indexPath, struct{}{})
	if err != nil {
		return err
	}
	htmlPath := filepath.Join(volumesPath, "index.html")
	if err := os.WriteFile(htmlPath, indexData, 0644); err != nil {
		return err
	}

	fmt.Println(T("generate-create-compose"))
	dataDir, _ := project.GetDataPath("prod")
	logsDir, _ := project.GetLogsPath("prod")

	composeData := GenerateData{
		DataDir:     dataDir,
		LogsDir:     logsDir,
		ServicePort: project.GetNumericPrefix(8000),
	}
	composePath := filepath.Join("maintenance", "compose.yaml")
	project.Compose, err = TemplateParse(composePath, composeData)
	if err != nil {
		return err
	}
	if err := project.SaveCompose(); err != nil {
		return err
	}

	return nil
}
