package agent

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"gopkg.in/ini.v1"
)

func OCISConfigPath(path string) string {
	configPath := GetToolsDir("data", "ocis", ".ocis", "config")
	if path == "" {
		return configPath
	}
	return filepath.Join(configPath, path)
}

func OCISTest(req *Request) bool {
	yamlPath := OCISConfigPath("ocis.yaml")
	if info, err := os.Stat(yamlPath); err == nil && info.Size() > 0 {
		return false
	}

	envPath := OCISConfigPath("ocis.env")
	if info, err := os.Stat(envPath); err == nil && info.Size() > 0 {
		return true
	}

	return false
}

func OCISHandler(req *Request, resp *Response) error {
	if OCISTest(req) == false {
		return nil
	}

	envPath := OCISConfigPath("ocis.env")
	envFile, err := ini.Load(envPath)
	if err != nil {
		return err
	}

	password := envFile.Section("").Key("OCIS_ADMIN_PASSWORD").String()
	resp.Sayf("OCIS password length is: %d", len(password))

	usr, err := user.Lookup("ocis")
	if err != nil {
		return fmt.Errorf("user lookup for ocis failed: %w", err)
	}
	uid, _ := strconv.Atoi(usr.Uid)
	gid, _ := strconv.Atoi(usr.Gid)

	args := []string{
		"init",
		"--config-path", OCISConfigPath(""),
		"--insecure", "yes",
		"--admin-password", password,
	}
	cmd := exec.Command(GetBinDir("ocis"), args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Credential: &syscall.Credential{
			Uid: uint32(uid),
			Gid: uint32(gid),
		},
	}

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ocis init failed: %w\n%s", err, output.String())
	}

	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		resp.Say(scanner.Text())
	}

	return nil
}
