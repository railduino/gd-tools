package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	SystemConfigName = "gd-tools-system.json"
	SystemIDsName    = "gd-tools-ids.json"
	SystemVarMount   = "/var/gd-tools"
	SystemDataRoot   = SystemVarMount + "/data"
	SystemLogsRoot   = SystemVarMount + "/logs"
)

type Mount struct {
	Provider   string `json:"provider"`   // e.g. "hetzner"
	Identifier string `json:"identifier"` // e.g. "123456789"
	Mountpoint string `json:"mountpoint"` // e.g. "/var/gd-tools"
}

type SystemIDs struct {
	ToolsUID  string `json:"tools_uid"`
	ToolsGID  string `json:"tools_gid"`
	DockerGID string `json:"docker_gid"`
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
	SystemIDs

	// webserver configuration
	ServeConfig

	// runtime data, like current path or flags
	CurrentPath string `json:"-"`
	DryRun      bool   `json:"-"`
	Progress    bool   `json:"-"`
	Upgrade     bool   `json:"-"`
}

func ReadSystemConfig(needIDs bool) (*SystemConfig, error) {
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
				systemConfig.SystemIDs = uidData
			}
		}
	}
	systemConfig.CurrentPath = currentPath

	if needIDs {
		if systemConfig.ToolsUID == "" || systemConfig.ToolsUID == "0" {
			msg := T("system-err-missing-ids")
			return nil, fmt.Errorf(msg)
		}
	}

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
