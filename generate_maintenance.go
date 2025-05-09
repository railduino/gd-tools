package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	RegisterProjectKind(generateMaintenance)
}

var generateMaintenance = &cli.Command{
	Name:        "maintenance",
	Usage:       T("generate-maintenance-usage"),
	Description: T("generate-maintenance-describe"),
	Flags:       []cli.Flag{},
	Action:      runGenerateMaintenance,
}

func runGenerateMaintenance(c *cli.Context) error {
	systemConfig, err := ReadSystemConfig(true)
	if err != nil {
		return err
	}

	project, err := GenerateBuildProject(c, "maintenance", true)
	if err != nil {
		return err
	}

	fmt.Println(T("generate-create-dir"))
	dataPath := filepath.Join(project.GetName(), "data")
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
	indexTmplPath := filepath.Join("maintenance", "index.html")
	indexData, err := TemplateParse(indexTmplPath, struct{}{})
	if err != nil {
		return err
	}
	htmlPath := filepath.Join(dataPath, "index.html")
	if err := os.WriteFile(htmlPath, indexData, 0644); err != nil {
		return err
	}

	fmt.Println(T("generate-create-compose"))
	composeData := GenerateData{
		DataDir:     filepath.Join(SystemDataRoot, project.GetName()),
		LogsDir:     filepath.Join(SystemLogsRoot, project.GetName()),
		ServicePort: project.PortStr,
		ToolsUID:    systemConfig.ToolsUID,
		DockerGID:   systemConfig.DockerGID,
	}
	composeTmplPath := filepath.Join("maintenance", "compose.yaml")
	project.Compose, err = TemplateParse(composeTmplPath, composeData)
	if err != nil {
		return err
	}
	if err := project.SaveCompose(); err != nil {
		return err
	}

	return nil
}
