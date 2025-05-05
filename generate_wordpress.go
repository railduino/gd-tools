package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

type WordPressTemplateData struct {
	Title     string `json:"title"`
	AdminUser string `json:"admin_user"`
	AdminMail string `json:"admin_mail"`

	SystemIDs
}

var wordPressFlagTitle = cli.StringFlag{
	Name:  "title",
	Usage: T("wordpress-flag-title"),
}

func init() {
	RegisterProjectKind(generateWordPress)
}

var generateWordPress = &cli.Command{
	Name:        "wordpress",
	Usage:       T("generate-wordpress-usage"),
	Description: T("generate-wordpress-describe"),
	Flags: []cli.Flag{
		&generateDependsFlag,
		&wordPressFlagTitle,
	},
	Action: runGenerateWordPress,
}

func runGenerateWordPress(c *cli.Context) error {
	args := c.Args().Slice()
	if len(args) < 1 {
		return fmt.Errorf(T("generate-err-missing-prefix"))
	}
	if len(args) < 2 {
		return fmt.Errorf(T("generate-err-missing-name"))
	}

	systemConfig, err := ReadSystemConfig(true)
	if err != nil {
		return err
	}

	project := Project{
		Prefix: args[0],
		Kind:   "wordpress",
		Name:   args[1],
	}
	if err := project.CheckConflict(true); err != nil {
		return err
	}

	fmt.Println(T("generate-create-dir"))
	if err := os.MkdirAll(project.GetName(), 0755); err != nil {
		return err
	}

	project.Config = Config{
		IsEnabled: false,
		DependsOn: []string{},
		Networks:  []string{},
	}
	for _, depend := range c.StringSlice("depends") {
		project.Config.DependsOn = append(project.Config.DependsOn, depend)
	}
	if err := project.SaveConfig(); err != nil {
		return err
	}

	title := c.String("title")
	if title == "" {
		title = fmt.Sprintf("%s @ %s", project.GetName(), systemConfig.DomainName)
	}

	composeData := WordPressTemplateData{
		Title:     title,
		AdminUser: fmt.Sprintf("admin@%s", systemConfig.DomainName),
		AdminMail: systemConfig.SysAdmin,
	}

	fmt.Println(T("generate-create-compose"))
	composePath := filepath.Join("wordpress", "compose.yaml")
	project.Compose, err = TemplateParse(composePath, composeData)
	if err != nil {
		return err
	}
	if err := project.SaveCompose(); err != nil {
		return err
	}

	return nil
}
