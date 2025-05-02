package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SystemConfigFile = "system.json"
)

type Mount struct {
	Provider   string `json:"provider"`   // eg. "hetzner"
	Identifier string `json:"identifier"` // eg. "123456789"
	Mountpoint string `json:"mountpoint"` // eg. "/var/gd-tools"
}

type SystemConfig struct {
	Version    string   `json:"version"`     // eg. v1.0.0
	TimeZone   string   `json:"time_zone"`   // eg. Europe/Berlin
	SwapSpace  int      `json:"swap_space"`  // size in GByte (or 0)
	HostName   string   `json:"host_name"`   // hostname (default FQDN)
	DomainName string   `json:"domain_name"` // derived from FQDN
	SshPort    string   `json:"ssh_port"`    // unsually 'OpenSSH'
	SysAdmin   string   `json:"sys_admin"`   // try to read from ~/.gitconfig
	Packages   []string `json:"packages"`    // Required DEB packages
	Mounts     []Mount  `json:"mounts"`      // Mounted filesystem (can grow)
}

func FileSystemRead() (*SystemConfig, error) {
	rootDir, err := GetProjectRoot(MainOnProd())
	if err != nil {
		return nil, err
	}

	systemFile := filepath.Join(rootDir, SystemConfigFile)
	systemData, err := os.ReadFile(systemFile)
	if err != nil {
		if os.IsNotExist(err) {
			msg := T("file-err-missing-system")
			return nil, fmt.Errorf(msg)
		}
		return nil, err
	}

	var systemConfig SystemConfig
	if err := json.Unmarshal(systemData, &systemConfig); err != nil {
		return nil, err
	}

	return &systemConfig, nil
}

func (sc *SystemConfig) Save() error {
	content, err := json.MarshalIndent(*sc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(SystemConfigFile, content, 0644); err != nil {
		return err
	}

	return nil
}
