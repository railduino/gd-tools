package php

import (
	"fmt"
	"path/filepath"
)

type TemplateData struct {
	PhpVersion string
}

var (
	PhpVersion string
	PhpEtcDir  string
)

func SetPhpVersion(version string) {
	PhpVersion = version
}

func GetPhpVersion() string {
	if PhpVersion == "" {
		PhpVersion = "8.3" // default for Ubuntu 24.4 noble
	}

	return PhpVersion
}

func GetPhpFpmService() string {
	return "php" + GetPhpVersion() + "-fpm"
}

func SetPhpEtcDir(path string) {
	if path == "" {
		path = "/etc/php/" + GetPhpVersion()
	}
	PhpEtcDir = path
}

func GetPhpEtcDir(paths ...string) string {
	if PhpEtcDir == "" {
		SetPhpEtcDir("")
	}
	if len(paths) == 0 {
		return PhpEtcDir
	}
	return filepath.Join(append([]string{PhpEtcDir}, paths...)...)
}

func GetPhpFpmPoolPath(number int, name string) string {
	filename := fmt.Sprintf("%02d-%s.conf", number, name)
	return GetPhpEtcDir("fpm", "pool.d", filename)
}

func GetTemplateData() TemplateData {
	return TemplateData{
		PhpVersion: GetPhpVersion(),
	}
}
