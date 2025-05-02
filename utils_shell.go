package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// ShellCmd runs exactly one command
func ShellCmd(dryRun bool, cmdStr string) error {
	cmd, err := shellPrepare(cmdStr)
	if err != nil {
		return err
	}

	if dryRun {
		fmt.Println(Tf("exec-dry-running", cmdStr))
		return nil
	}

	fmt.Println(Tf("exec-now-running", cmdStr))
	cmd.Env = append(os.Environ(), "LANG=C")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func ShellCmds(dryRun bool, cmdList []string) error {
	for _, cmdStr := range cmdList {
		if err := ShellCmd(dryRun, cmdStr); err != nil {
			return err
		}
	}
	return nil
}

func ShellMatch(cmdStr string, pattern string) (bool, error) {
	cmd, err := shellPrepare(cmdStr)
	if err != nil {
		return false, err
	}

	cmd.Env = append(os.Environ(), "LANG=C")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, err
	}

	matched, err := regexp.MatchString(pattern, string(out))
	if err != nil {
		return false, err
	}

	return matched, nil
}

func ShellEditor(path string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim"
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func SystemService(dryRun bool, serviceName string) error {
	statusCmd := exec.Command("systemctl", "is-active", serviceName)
	if err := statusCmd.Run(); err == nil {
		fmt.Printf("- Service %s ist bereits aktiv\n", serviceName)
		return nil
	}

	cmd := fmt.Sprintf("systemctl enable --now %s", serviceName)
	return ShellCmd(dryRun, cmd)
}

func shellPrepare(cmdStr string) (*exec.Cmd, error) {
	if cmdStr == "" {
		return nil, fmt.Errorf(T("exec-err-missing"))
	}

	args := strings.Fields(cmdStr)
	if len(args) < 1 {
		return nil, fmt.Errorf(Tf("exec-err-invalid", cmdStr))
	}

	return exec.Command(args[0], args[1:]...), nil
}
