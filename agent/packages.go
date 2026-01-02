package agent

import (
	"fmt"
	"os"
	"path/filepath"
)

// N.B. Some years from now, read the version from /etc/os-release
const (
	Ubuntu = "noble"
)

// PackagesTest checks if there is work to be done
func PackagesTest(req *Request) bool {
	marker := GetRootDir(FirstRunMarker)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		return true
	}
	return req != nil && len(req.Packages) > 0
}

func PackagesHandler(req *Request, resp *Response) error {
	if req == nil || resp == nil {
		return nil
	}

	marker := GetRootDir(FirstRunMarker)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		return fmt.Errorf("missing %s - please deploy 01-bootstrap", marker)
	}

	if len(req.Packages) == 0 && req.Upgrade == false {
		return nil
	}

	if _, err := RunCommand("apt-get", "update", "--quiet"); err != nil {
		return err
	}

	if req.Upgrade {
		if _, err := RunCommand("apt-get", "upgrade", "--yes", "--quiet"); err != nil {
			return err
		}
	}

	preconfCmds := []string{
		"postfix postfix/main_mailer_type select No configuration",
	}
	for _, preconf := range preconfCmds {
		cmd := fmt.Sprintf("echo '%s' | debconf-set-selections", preconf)
		if _, err := RunShell([]string{cmd}); err != nil {
			return err
		}
	}

	count := 0
	for _, pkg := range req.Packages {
		if _, err := RunCommand("dpkg", "--verify", pkg); err == nil {
			count++
			continue
		}

		if _, err := RunCommand("apt-get", "--quiet", "--yes", "install", pkg); err != nil {
			return err
		}
		resp.Sayf(" - successfully installed %s", pkg)
	}

	if err := EnsureDockerComposeSymlink(); err != nil {
		return err
	}

	resp.AddService("ssh")

	if count > 0 {
		resp.Sayf("✅ %d packages installed", count)
	}

	return nil
}

func EnsureDockerComposeSymlink() error {
	src := "/usr/libexec/docker/cli-plugins/docker-compose"
	dst := "/usr/local/bin/docker-compose"

	if _, err := os.Stat(src); err != nil {
		if os.IsNotExist(err) {
			_ = os.Remove(dst)
			return nil // docker-compose plugin not installed
		}
		return err
	}

	// Ensure target directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// Remove existing file/symlink if present
	_ = os.Remove(dst)

	return os.Symlink(src, dst)
}
