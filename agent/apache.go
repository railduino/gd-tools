package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func GetApacheToolsDir(paths ...string) string {
	toolsDir := GetToolsDir("data", "apache")
	if len(paths) == 0 {
		return toolsDir
	}
	return filepath.Join(append([]string{toolsDir}, paths...)...)
}

func GetApacheEtcDir(paths ...string) string {
	etcDir := GetEtcDir("apache2")
	if len(paths) == 0 {
		return etcDir
	}
	return filepath.Join(append([]string{etcDir}, paths...)...)
}

func (file *File) IsApacheAvailable() (string, string, bool) {
	for _, kind := range []string{"conf", "mods", "sites"} {
		dir := GetApacheEtcDir(kind + "-available/")
		if module, ok := strings.CutPrefix(file.Path, dir); ok {
			return kind, module, true
		}
	}

	return "", "", false
}

func (file *File) ApacheEnable(kind, module string) error {
	src := GetApacheEtcDir(kind+"-available", module)
	dest := GetApacheEtcDir(kind+"-enabled", module)

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("failed to create symlink for %s: %w", module, err)
	}
	if _, err := os.Lstat(dest); err == nil {
		return nil
	}

	relSrc := filepath.Join("..", kind+"-available", module)
	if err := os.Symlink(relSrc, dest); err != nil {
		return fmt.Errorf("failed to create symlink for %s: %w", module, err)
	}

	return nil
}

func (file *File) ApacheDisable(kind, module string) error {
	enabled := GetApacheEtcDir(kind + "-enabled")

	for _, ext := range []string{".load", ".conf"} {
		dest := filepath.Join(enabled, module+ext)

		if _, err := os.Lstat(dest); err == nil {
			if err := os.Remove(dest); err != nil {
				return fmt.Errorf("failed to remove symlink for %s: %w", module, err)
			}
		}
	}

	return nil
}
