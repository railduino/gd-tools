package agent

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RunShell executes a list of shell commands on the local system.
// Commands are joined with && and run as a single shell script.
// Returns an error if any command fails.
func RunShell(commands []string) ([]byte, error) {
	if len(commands) == 0 {
		return nil, fmt.Errorf("no local commands provided")
	}
	script := "set -e; " + strings.Join(commands, " && ")

	cmd := exec.Command("sh", "-c", script)
	cmd.Env = append(os.Environ(), "LANG=C")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("shell failed: %w\nOutput:\n%s", err, output)
	}

	return output, nil
}

// RunCommand executes a single shell command.
func RunCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "LANG=C")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %w\nOutput:\n%s", err, output)
	}
	return output, nil
}

func StartService(service string) (string, error) {
	cmds := []string{
		"systemctl daemon-reload",
		"systemctl enable " + service,
		"systemctl reset-failed " + service,
		"systemctl restart " + service,
	}

	if _, err := RunShell(cmds); err != nil {
		return "", fmt.Errorf("failed to start service %s: %w", service, err)
	}

	return fmt.Sprintf("✅ service %s was (re)started", service), nil
}
