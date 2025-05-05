package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SystemConfigName = "gd-tools-config.json"
	SystemIDsName    = "gd-tools-ids.json"
)

type Mount struct {
	Provider   string `json:"provider"`   // e.g. "hetzner"
	Identifier string `json:"identifier"` // e.g. "123456789"
	Mountpoint string `json:"mountpoint"` // e.g. "/var/gd-tools"
}

type SystemIDs struct {
	ToolsUID  int `json:"tools_uid"`
	ToolsGID  int `json:"tools_gid"`
	DockerGID int `json:"docker_gid"`
}

type SystemConfig struct {
	Version    string   `json:"version"`     // e.g. v1.0.0
	TimeZone   string   `json:"time_zone"`   // e.g. Europe/Berlin
	HostName   string   `json:"host_name"`   // hostname (default FQDN)
	DomainName string   `json:"domain_name"` // derived from FQDN
	SwapSpace  int      `json:"swap_space"`  // size in GByte (or 0)
	SysAdmin   string   `json:"sys_admin"`   // try to read from ~/.gitconfig
	Packages   []string `json:"packages"`    // Required DEB packages
	Mounts     []Mount  `json:"mounts"`      // Mounted filesystem (can grow)

	// container uid/gid - fetch after deployment
	ToolsUID  int `json:"tools_uid"`
	ToolsGID  int `json:"tools_gid"`
	DockerGID int `json:"docker_gid"`

	// runtime data, like current path or flags
	CurrentPath string `json:"-"`
	DryRun      bool   `json:"-"`
	Progress    bool   `json:"-"`
	Upgrade     bool   `json:"-"`
}

func ReadSystemConfig() (*SystemConfig, error) {
	currentPath, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	systemConfigPath := SystemConfigName
	if CheckEnv("prod") {
		systemConfigPath = filepath.Join("/etc", SystemConfigName)
	}

	content, err := os.ReadFile(systemConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			msg := T("system-err-missing-file")
			return nil, fmt.Errorf(msg)
		}
		return nil, err
	}

	var systemConfig SystemConfig
	if err := json.Unmarshal(content, &systemConfig); err != nil {
		return nil, err
	}

	if CheckEnv("dev") {
		uidPath := filepath.Join("letsencrypt", SystemIDsName)
		if content, err := os.ReadFile(uidPath); err == nil {
			var uidData SystemIDs
			if err := json.Unmarshal(content, &uidData); err == nil {
				systemConfig.ToolsUID = uidData.ToolsUID
				systemConfig.ToolsGID = uidData.ToolsGID
				systemConfig.DockerGID = uidData.DockerGID
			}
		}
	}
	systemConfig.CurrentPath = currentPath

	return &systemConfig, nil
}

func (sc *SystemConfig) Save() error {
	content, err := json.MarshalIndent(*sc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(SystemConfigName, content, 0644); err != nil {
		return err
	}

	return nil
}
