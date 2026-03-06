package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/ini.v1"
)

const (
	BasicsName = "basics.json"

	DefaultTimeZone = "Europe/Berlin"
	DefaultCompany  = "My Company"
	DefaultRegTTL   = 3600
	DefaultDMARC    = "v=DMARC1; p=quarantine; pct=100; adkim=s; aspf=s"
)

type Basics struct {
	Company  string `json:"company"`
	SysAdmin string `json:"sys_admin"`
	TimeZone string `json:"time_zone"`
	Language string `json:"language"`
	Region   string `json:"region"`
	RegTTL   int    `json:"reg_ttl"`
	HelpURL  string `json:"help_url"`
	DMARC    string `json:"dmarc"`
}

func (basics *Basics) Locale() string {
	return basics.Language + "_" + basics.Region
}

func ReadBasics() (*Basics, error) {
	content, err := os.ReadFile(BasicsName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("missing %s - are we in $(GD_TOOLS_BASE)?", BasicsName)
		}
		return nil, fmt.Errorf("failed to read %s: %w", BasicsName, err)
	}

	var basics Basics
	if err := json.Unmarshal(content, &basics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", BasicsName, err)
	}

	if basics.Company == DefaultCompany {
		return nil, fmt.Errorf("%s has default values - please edit", BasicsName)
	}

	if basics.DMARC == "" {
		basics.DMARC = DefaultDMARC
	}

	return &basics, nil
}

func GetBasics() (*Basics, error) {
	path := filepath.Join("..", BasicsName)

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("missing %s - are we in the correct dir?", path)
	}

	var basics Basics
	if err := json.Unmarshal(content, &basics); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", path, err)
	}

	if basics.DMARC == "" {
		basics.DMARC = DefaultDMARC
	}

	return &basics, nil
}

func (basics *Basics) Save() error {
	content, err := json.MarshalIndent(basics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", BasicsName, err)
	}

	existing, err := os.ReadFile(BasicsName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(BasicsName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", BasicsName, err)
	}

	return nil
}

func GetTimeZone() string {
	target, err := os.Readlink("/etc/localtime")
	if err != nil {
		return DefaultTimeZone
	}

	rel, err := filepath.Rel("/usr/share/zoneinfo", target)
	if err != nil {
		return DefaultTimeZone
	}

	return rel
}

func GetSysAdmin() string {
	if homeDir, err := os.UserHomeDir(); err == nil {
		gitConfigPath := filepath.Join(homeDir, ".gitconfig")
		if gitConfig, err := ini.Load(gitConfigPath); err == nil {
			return gitConfig.Section("user").Key("email").String()
		}
	}

	return "admin@" + DefaultCompany
}
