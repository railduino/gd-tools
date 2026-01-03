package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// N.B. Some years from now, read the version from /etc/os-release
const (
	Ubuntu = "noble"
)

type proStatusJSON struct {
	Attached *bool `json:"attached"`
}

// PackagesTest checks if there is work to be done
func PackagesTest(req *Request) bool {
	marker := GetRootDir(FirstRunMarker)
	if _, err := os.Stat(marker); os.IsNotExist(err) {
		return true
	}
	if req == nil {
		return false
	}
	return len(req.Packages) > 0 || req.Upgrade || req.UbuntuPro != ""
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

	if err := handleUbuntuPro(req, resp); err != nil {
		return err
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

	// N.B. Docker is not installed by default. It must be requested by an app.
	// But the official Docker Repo is already prepared.
	// And if Docker is installed, make sure there is a docker-compose command in PATH
	if err := ensureDockerComposeSymlink(); err != nil {
		return err
	}

	resp.AddService("ssh")

	if count > 0 {
		resp.Sayf("✅ %d packages installed", count)
	}

	return nil
}

func handleUbuntuPro(req *Request, resp *Response) error {
	if req == nil || resp == nil {
		return nil
	}

	// Ubuntu Pro not desired → do not introduce Ubuntu dependency
	if req.UbuntuPro == "" {
		if _, err := exec.LookPath("pro"); err == nil {
			attached, _ := isUbuntuProAttached()
			if attached {
				if _, err := RunCommand("pro", "detach", "--assume-yes"); err != nil {
					return err
				}
			}
		}
		resp.Sayf("⭕ No Ubuntu Pro")
		return nil
	}

	// Ubuntu Pro explicitly desired → now Ubuntu assumptions are allowed
	if err := ensureUbuntuProClientInstalled(resp); err != nil {
		return err
	}

	attached, _ := isUbuntuProAttached()
	if attached {
		resp.Sayf("✅ Ubuntu Pro is already attached")
		return nil
	}

	if _, err := RunCommand("pro", "attach", req.UbuntuPro); err != nil {
		return err
	}
	resp.Sayf("✅ Ubuntu Pro has been attached")

	return nil
}

func isUbuntuProAttached() (bool, error) {
	out, err := RunCommand("pro", "status", "--format", "json")
	if err != nil {
		// If pro is not available or fails, treat as not attached.
		return false, nil
	}

	var st proStatusJSON
	if err := json.Unmarshal([]byte(out), &st); err != nil {
		// JSON is marked experimental; be conservative.
		return false, nil
	}
	if st.Attached == nil {
		return false, nil
	}

	return *st.Attached, nil
}

func ensureUbuntuProClientInstalled(resp *Response) error {
	if _, err := exec.LookPath("pro"); err == nil {
		return nil
	}

	// We need package metadata before installing the client
	if _, err := RunCommand("apt-get", "update", "--quiet"); err != nil {
		return err

	}

	// Install ubuntu-pro-client.
	if _, err := RunCommand("apt-get", "--quiet", "--yes", "install", "ubuntu-pro-client"); err != nil {
		return err
	}
	resp.Sayf(" - successfully installed ubuntu-pro-client")

	return nil
}

func ensureDockerComposeSymlink() error {
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
