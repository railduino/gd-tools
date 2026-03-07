package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/railduino/gd-tools/php"
)

const (
	NextcloudName  = "nextcloud.json"
	FirstNextcloud = 21
	LastNextcloud  = 24
)

type Nextcloud struct {
	Name       string   `json:"name"`
	Number     int      `json:"number"`
	Language   string   `json:"language"`
	Region     string   `json:"region"`
	HostName   string   `json:"host_name"`
	DomainName string   `json:"domain_name"`
	Version    string   `json:"version"`
	ServerFQDN string   `json:"server_fqdn"`
	Subdir     string   `json:"subdir"`
	Password   string   `json:"password"`
	InstanceID string   `json:"instance_id"`
	Salt       string   `json:"salt"`
	Secret     string   `json:"secret"`
	AdminEmail string   `json:"admin_email"`
	AppList    []string `json:"app_list,omitempty"`

	Download *Download `json:"-"`
}

type NextcloudList struct {
	Entries []*Nextcloud `json:"entries"`
}

func (nc *Nextcloud) FQDN() string {
	return nc.HostName + "." + nc.DomainName
}

func (nc *Nextcloud) RootDir() string {
	return GetToolsDir("data", "nextcloud", nc.Name)
}

func (nc *Nextcloud) SocketPath() string {
	name := fmt.Sprintf("php%s-nextcloud-%s.sock", php.GetPhpVersion(), nc.Name)
	return filepath.Join("/run/php", name)
}

func (nc *Nextcloud) ConfigPath() string {
	return filepath.Join(nc.RootDir(), "config.json")
}

func (nc *Nextcloud) BaseDir(paths ...string) string {
	baseDir := filepath.Join(nc.RootDir(), "nextcloud")
	if len(paths) == 0 {
		return baseDir
	}
	return filepath.Join(append([]string{baseDir}, paths...)...)
}

func (nc *Nextcloud) SlashSubdir() string {
	s := strings.Trim(nc.Subdir, " /")
	if s == "" {
		return ""
	}
	return "/" + s
}

func (nc *Nextcloud) DataDir(paths ...string) string {
	dataDir := filepath.Join(nc.RootDir(), "data")
	if len(paths) == 0 {
		return dataDir
	}
	return filepath.Join(append([]string{dataDir}, paths...)...)
}

func (nc *Nextcloud) LogsDir(paths ...string) string {
	logsDir := GetToolsDir("logs", "nextcloud", nc.Name)
	if len(paths) == 0 {
		return logsDir
	}
	return filepath.Join(append([]string{logsDir}, paths...)...)
}

func (nc *Nextcloud) CronPath() string {
	return GetEtcDir("cron.d", "nextcloud_"+nc.Name)
}

func (nc *Nextcloud) VhostPath() string {
	name := fmt.Sprintf("%02d-%s.conf", nc.Number, nc.FQDN())
	return GetApacheEtcDir("sites-available", name)
}

func (nc *Nextcloud) HookPath() string {
	name := "backup-pre-nextcloud-" + nc.Name
	return GetToolsDir("data", "hooks", name)
}

func (nc *Nextcloud) CertDir() string {
	return GetToolsDir("data", "certs", nc.FQDN())
}

// the following functions are on Dev
func LoadNextcloudList(update *Nextcloud) (*NextcloudList, error) {
	var list NextcloudList

	content, err := os.ReadFile(NextcloudName)
	if err != nil {
		if os.IsNotExist(err) {
			return &list, nil
		}
		return nil, fmt.Errorf("failed to read %s: %w", NextcloudName, err)
	}

	if err := json.Unmarshal(content, &list); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", NextcloudName, err)
	}

	for index, _ := range list.Entries {
		entry := list.Entries[index]
		if entry.Name == "" {
			return nil, fmt.Errorf("found Nextcloud without Name")
		}
		if entry.Number < FirstNextcloud || entry.Number > LastNextcloud {
			return nil, fmt.Errorf("invalid number %d for Nextcloud %s", entry.Number, entry.Name)
		}
		if entry.Language == "" {
			entry.Language = GetLanguage()
		}
		if entry.Region == "" {
			entry.Region = GetRegion()
		}
		if entry.HostName == "" {
			return nil, fmt.Errorf("missing HostName for Nextcloud %s", entry.Name)
		}
		if entry.DomainName == "" {
			return nil, fmt.Errorf("missing DomainName for Nextcloud %s", entry.Name)
		}

		if update != nil && update.Password != "" {
			entry.Password = update.Password
		}
		if entry.Password == "" {
			return nil, fmt.Errorf("missing Password for Nextcloud %s", entry.Name)
		}

		if update != nil && update.InstanceID != "" {
			entry.InstanceID = update.InstanceID
		}
		if entry.InstanceID == "" {
			return nil, fmt.Errorf("missing InstanceID for Nextcloud %s", entry.Name)
		}

		if update != nil && update.Salt != "" {
			entry.Salt = update.Salt
		}
		if entry.Salt == "" {
			return nil, fmt.Errorf("missing Salt for Nextcloud %s", entry.Name)
		}

		if update != nil && update.Secret != "" {
			entry.Secret = update.Secret
		}
		if entry.Secret == "" {
			return nil, fmt.Errorf("missing Secret for Nextcloud %s", entry.Name)
		}

		if update != nil && update.AdminEmail != "" {
			entry.AdminEmail = update.AdminEmail
		}
		if entry.AdminEmail == "" {
			return nil, fmt.Errorf("missing AdminEmail for Nextcloud %s", entry.Name)
		}
	}

	if err := list.Save(); err != nil {
		return nil, err
	}

	if update != nil {
		return nil, nil
	}

	return &list, nil
}

func (list *NextcloudList) Save() error {
	sort.Slice(list.Entries, func(i, j int) bool {
		return list.Entries[i].Name < list.Entries[j].Name
	})

	content, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", NextcloudName, err)
	}

	existing, err := os.ReadFile(NextcloudName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(NextcloudName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", NextcloudName, err)
	}

	return nil
}
