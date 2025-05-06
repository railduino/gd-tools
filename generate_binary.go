package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

func init() {
	RegisterProjectKind(generateBinary)
}

var generateBinary = &cli.Command{
	Name:        "binary",
	Usage:       T("install-binary-usage"),
	Description: T("install-binary-describe"),
	Flags:       []cli.Flag{},
	Action:      runGenerateBinary,
}

func runGenerateBinary(c *cli.Context) error {
	systemConfig, err := ReadSystemConfig(true)
	if err != nil {
		return err
	}

	project, err := GenerateBuildProject(c, "binary", false)
	if err != nil {
		return err
	}

	fmt.Println(T("generate-create-dirs"))
	dataPath := filepath.Join(project.GetName(), "data")
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return err
	}
	binPath := filepath.Join(project.GetName(), "bin")
	if err := os.MkdirAll(binPath, 0755); err != nil {
		return err
	}
	keepPath := filepath.Join(project.GetName(), "bin", ".keep")
	if err := os.WriteFile(keepPath, []byte(""), 0644); err != nil {
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

	fmt.Println(T("generate-create-compose"))
	composeData := GenerateData{
		DataDir:     filepath.Join(SystemDataRoot, project.GetName()),
		LogsDir:     filepath.Join(SystemLogsRoot, project.GetName()),
		ServicePort: project.PortStr,
		ToolsUID:    systemConfig.ToolsUID,
		DockerGID:   systemConfig.DockerGID,
	}
	composeTmplPath := filepath.Join("binary", "compose.yaml")
	project.Compose, err = TemplateParse(composeTmplPath, composeData)
	if err != nil {
		return err
	}
	if err := project.SaveCompose(); err != nil {
		return err
	}

	return nil
}
