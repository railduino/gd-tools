package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	ProdProjectRoot = "/etc/gd-tools"
	ProdDataRoot    = "/var/gd-tools"
)

type Config struct {
	IsEnabled bool     `json:"is_enabled"`
	DependsOn []string `json:"depends_on"`
	Networks  []string `json:"networks"`
}

type Project struct {
	Prefix  string
	Kind    string
	Name    string
	Config  Config
	Compose []byte
}

func GetProjectRoot(prod bool) (string, error) {
	if prod {
		if MainOnProd() {
			if env := os.Getenv("GD_PROJECT_ROOT"); env != "" {
				return env, nil
			}
		}
		return ProdProjectRoot, nil
	}

	curDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return curDir, nil
}

func GetDataRoot(prod bool, subdir string) (string, error) {
	if prod {
		baseDir := ProdDataRoot
		if MainOnProd() {
			if env := os.Getenv("GD_DATA_ROOT"); env != "" {
				baseDir = env
			}
		}
		if subdir == "" {
			return baseDir, nil
		}
		return filepath.Join(baseDir, subdir), nil
	}

	if _, err := os.Getwd(); err != nil {
		return "", err
	}

	return subdir, nil
}

func ProjectLoadAll(prod bool) ([]*Project, error) {
	rootDir, err := GetProjectRoot(prod)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return nil, err
	}

	var projects []*Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		parts := strings.SplitN(entry.Name(), "-", 3)
		if len(parts) < 2 {
			continue
		}
		if len(parts) < 3 {
			parts = append(parts, "")
		}

		p := &Project{
			Prefix: parts[0],
			Kind:   parts[1],
			Name:   parts[2],
		}

		projects = append(projects, p)
	}

	return projects, nil
}

func SortProjectsAscending(projects []*Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})
}

func SortProjectsDescending(projects []*Project) {
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name > projects[j].Name
	})
}

func (p *Project) GetName() string {
	parts := []string{p.Prefix, p.Kind}
	if p.Name != "" {
		parts = append(parts, p.Name)
	}

	return strings.Join(parts, "-")
}

func (p *Project) GetPath(prod bool) (string, error) {
	rootDir, err := GetProjectRoot(prod)
	if err != nil {
		return "", err
	}

	return filepath.Join(rootDir, p.GetName()), nil
}

func (p *Project) GetDataPath(prod bool) (string, error) {
	rootDir, err := GetDataRoot(prod, "volumes")
	if err != nil {
		return "", err
	}

	return filepath.Join(rootDir, p.GetName()), nil
}

func (p *Project) GetLogsPath(prod bool) (string, error) {
	rootDir, err := GetDataRoot(prod, "logs")
	if err != nil {
		return "", err
	}

	return filepath.Join(rootDir, p.GetName()), nil
}

func (p *Project) CheckConflict(unique bool) error {
	projects, err := ProjectLoadAll(false)
	if err != nil {
		return err
	}

	for _, check := range projects {
		if check.Prefix == p.Prefix && check.Kind == p.Kind && check.Name == p.Name {
			msg := T("install-err-project-exist")
			return fmt.Errorf(msg)
		}

		if unique && check.Kind == p.Kind {
			msg := T("install-err-unique-exist")
			return fmt.Errorf(msg)
		}
	}

	return nil
}

func (p *Project) ConfigPath() string {
	projectDir, _ := p.GetPath(MainOnProd())

	return filepath.Join(projectDir, "config.json")
}

func (p *Project) LoadConfig() error {
	content, err := os.ReadFile(p.ConfigPath())
	if err != nil {
		return err
	}

	if err := json.Unmarshal(content, &p.Config); err != nil {
		return err
	}

	return nil
}

func (p *Project) SaveConfig() error {
	content, err := json.MarshalIndent(p.Config, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(p.ConfigPath(), content, 0644); err != nil {
		return err
	}

	return nil
}

func (p *Project) ComposePath() string {
	projectDir, _ := p.GetPath(MainOnProd())

	return filepath.Join(projectDir, "compose.yaml")
}

func (p *Project) SaveCompose() error {
	if err := os.WriteFile(p.ComposePath(), p.Compose, 0644); err != nil {
		return err
	}

	return nil
}

/* TODO
func (p *Project) CheckDependencies(all []*Project) error {
	for _, depName := range p.Config.DependsOn {
		dep, _ := ProjectLoad(depName, false)
		if dep == nil {
			return fmt.Errorf("dependency %q not found for %s", depName, p.Name)
		}
		if !dep.Config.IsValid || !dep.Config.IsEnabled {
			return fmt.Errorf("dependency %q for %s is not valid/enabled", depName, p.Name)
		}
	}
	return nil
}
*/
