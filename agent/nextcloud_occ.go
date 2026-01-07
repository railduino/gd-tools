package agent

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

const (
	OCC_Install    = "maintenance:install"
	OCC_Config_Get = "config:system:get"
	OCC_Config_Set = "config:system:set"
)

type NextConfigEntry struct {
	Key   string
	Typ   string
	Value string
}

// NextcloudTest checks if there is work to be done
func NextcloudTest(req *Request) bool {
	return req != nil && req.Nextcloud != nil
}

func NextcloudHandler(req *Request, resp *Response) error {
	if req == nil || req.Nextcloud == nil || resp == nil {
		return nil
	}
	nc := req.Nextcloud

	path := nc.ConfigPath()
	if _, err := os.ReadFile(path); err != nil {
		return fmt.Errorf("failed to read Nextcloud '%s' config: %w", path, err)
	}

	if req.NextConf == OCC_Install {
		config := nc.BaseDir("config", "config.php")
		if info, err := os.Stat(config); err == nil && info.Size() > 0 {
			resp.Sayf("✅ Nextcloud '%s' installed", nc.Name)
			return nil
		}
		if err := nc.Install(resp); err != nil {
			return err
		}
	} else {
		if err := nc.Configure(resp, req.NextConf); err != nil {
			return err
		}
	}

	return nil
}

func (nc *Nextcloud) Install(resp *Response) error {
	resp.Sayf("installing Nextcloud '%s'", nc.Name)

	if err := ChownR(nc.BaseDir(), "www-data:www-data"); err != nil {
		return fmt.Errorf("chown failed for %s: %w", nc.BaseDir(), err)
	}
	resp.Sayf("Nextcloud '%s' belongs to www-data", nc.Name)

	resp.Sayf("running Nextcloud installer for '%s' - please be patient ...", nc.Name)
	if _, err := nc.RunOCC(resp, OCC_Install,
		"--database", "mysql",
		"--database-host", "localhost",
		"--database-port", "3306",
		"--database-name", "nc_"+nc.Name,
		"--database-user", "nc_"+nc.Name,
		"--database-pass", nc.Password,
		"--admin-user", "admin",
		"--admin-pass", nc.Password,
		"--admin-email", nc.AdminEmail,
		"--data-dir", nc.DataDir(),
	); err != nil {
		return err
	}
	_ = os.Remove(nc.BaseDir("config", "CAN_INSTALL"))

	return nil
}

func (nc *Nextcloud) Configure(resp *Response, key string) error {
	if resp == nil {
		return fmt.Errorf("missing response in Configure()")
	}

	var entry *NextConfigEntry
	for i, e := range nc.GetConfigList() {
		if e.Key == key {
			entry = &nc.GetConfigList()[i]
			break
		}
	}
	if entry == nil {
		return fmt.Errorf("config key not found: %s", key)
	}

	var value string
	var err error
	keyParts := strings.SplitN(key, ":", 2)
	if len(keyParts) == 2 {
		value, err = nc.RunOCC(resp, OCC_Config_Get, keyParts[0], keyParts[1])
	} else {
		value, err = nc.RunOCC(resp, OCC_Config_Get, key)
	}
	if err != nil {
		value = ""
		// return fmt.Errorf("config:system:get %s failed: %w", key, err)
	}
	if value == entry.Value {
		resp.Sayf("✅ %s (%s)", key, entry.Value)
		return nil
	}

	var args []string
	if len(keyParts) == 2 {
		args = append(args, keyParts[0], keyParts[1])
	} else {
		args = append(args, key)
	}
	args = append(args, "--type", entry.Typ, "--value", entry.Value)
	if _, err := nc.RunOCC(resp, OCC_Config_Set, args...); err != nil {
		return fmt.Errorf("set config %s to '%q' failed: %w", key, args, err)
	}
	resp.Sayf("updated config %s ==> '%s'", key, entry.Value)

	if nc.Subdir != "" {
		if _, err := nc.RunOCC(resp, "maintenance:update:htaccess"); err != nil {
			return fmt.Errorf("update:htaccess failed: %w", err)
		}
	}

	return nil
}

func (nc *Nextcloud) GetConfigList() []NextConfigEntry {
	adminParts := strings.SplitN(nc.AdminEmail, "@", 2)
	mailFrom := adminParts[0]
	mailDomain := ""
	if len(adminParts) > 1 {
		mailDomain = adminParts[1]
	}

	configEntries := []NextConfigEntry{
		{"instanceid", "string", nc.InstanceID},
		{"datadirectory", "string", nc.DataDir()},
		{"overwriteprotocol", "string", "https"},
		{"overwrite.cli.url", "string", "https://" + nc.FQDN() + nc.SlashSubdir()},
		{"trusted_domains:0", "string", nc.FQDN()},
		{"trusted_proxies:0", "string", "127.0.0.1"},
		{"trusted_proxies:1", "string", "::1"},
		{"trusted_domains:1", "string", nc.ServerFQDN},
		{"passwordsalt", "string", nc.Salt},
		{"secret", "string", nc.Secret},
		{"default_language", "string", nc.Language},
		{"default_phone_region", "string", nc.Region},
		{"mysql.utf8mb4", "boolean", "true"},
		{"memcache.local", "string", "\\OC\\Memcache\\Redis"},
		{"memcache.locking", "string", "\\OC\\Memcache\\Redis"},
		{"versions_retention_obligation", "string", "0"},
		{"maintenance_window_start", "integer", "1"},
		{"simpleSignUpLink.shown", "boolean", "false"},
		{"mail_from_address", "string", mailFrom},
		{"mail_domain", "string", mailDomain},
		{"mail_smtpmode", "string", "sendmail"},
		{"mail_sendmailmode", "string", "smtp"},
	}

	if nc.Subdir != "" {
		configEntries = append(configEntries,
			NextConfigEntry{"htaccess.RewriteBase", "string", nc.SlashSubdir()},
			NextConfigEntry{"overwritewebroot", "string", nc.SlashSubdir()},
		)
	}

	return configEntries
}

func (nc *Nextcloud) RunOCC(resp *Response, cmd string, args ...string) (string, error) {
	if resp == nil {
		return "", fmt.Errorf("missing resonse in RunOCC()")
	}

	usr, err := user.Lookup("www-data")
	if err != nil {
		return "", fmt.Errorf("user lookup for www-data failed: %w", err)
	}
	uid, _ := strconv.Atoi(usr.Uid)
	gid, _ := strconv.Atoi(usr.Gid)

	occPath := nc.BaseDir("occ")
	occArgs := []string{"-f", occPath}
	if cmd != "" {
		occArgs = append(occArgs, cmd)
		occArgs = append(occArgs, args...)
	}
	command := exec.Command("/usr/bin/php", occArgs...)
	command.Dir = nc.BaseDir()
	command.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output

	if err := command.Run(); err != nil {
		out := output.String()
		// Treat missing config keys as "empty" for idempotent setup runs
		if cmd == "config:system:get" && strings.Contains(out, "config key not found") {
			return "", nil
		}
		return "", fmt.Errorf("RunOCC failed: %w\n%s", err, output.String())
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) == 1 {
		return lines[0], nil
	}
	for _, line := range lines {
		resp.Say(line)
	}

	return "", nil
}
