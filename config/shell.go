package config

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/railduino/gd-tools/agent"
)

func (cfg *Config) LocalScript(commands []string) ([]byte, error) {
	if len(commands) == 0 {
		return nil, fmt.Errorf("no local commands provided")
	}
	script := "set -e; " + strings.Join(commands, " && ")

	cfg.Debugf("running '%s'", script)
	if cfg.Dry {
		return nil, nil
	}

	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(), "LANG=C")
	for _, env := range cfg.CmdEnv {
		cmd.Env = append(os.Environ(), env)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("shell failed: %w\nOutput:\n%s", err, output)
	}

	return output, nil
}

func (cfg *Config) LocalCommand(name string, args ...string) ([]byte, error) {
	cfg.Sayf("run '%s %s'", name, strings.Join(args, " "))
	if cfg.Dry {
		return nil, nil
	}

	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "LANG=C")
	for _, env := range cfg.CmdEnv {
		cmd.Env = append(os.Environ(), env)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nOutput:\n%s", err, output)
	}
	return output, nil
}

func (cfg *Config) CheckRemote(task string) bool {
	cmd := exec.Command("ssh", cfg.RootUser(), task)
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (cfg *Config) RemoteScript(commands []string) error {
	if len(commands) == 0 {
		return fmt.Errorf("no remote commands provided")
	}

	script := strings.Join(commands, " && ")
	return cfg.RemoteCmd(script)
}

func (cfg *Config) RemoteCmd(command string) error {
	rootUser := cfg.RootUser()

	cfg.Sayf("ssh %s %q", rootUser, command)
	if cfg.Dry {
		return nil
	}

	cmd := exec.Command("ssh", rootUser, command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ssh to %s failed: %w\nOutput:\n%s", rootUser, err, output)
	}

	return nil
}

func (cfg *Config) PushCerts() {
	if info, err := os.Stat("acme-certs"); err != nil || !info.IsDir() {
		return
	}

	dataPath := agent.GetToolsDir("data")
	if ok := cfg.CheckRemote("test -d " + dataPath); !ok {
		return
	}
	certPath := agent.GetToolsDir("data", "certs")

	if _, err := cfg.LocalCommand(
		"rsync",
		cfg.RsyncFlags(),
		"--chown=root:root",
		"--delete",
		"acme-certs/",
		cfg.RootUser()+":"+certPath,
	); err != nil {
		cfg.Sayf("failed to push ACME certs: %s (ignored)", err.Error())
	}
}
