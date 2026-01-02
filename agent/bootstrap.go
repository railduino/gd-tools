package agent

import (
	"fmt"
	"os"
	"strings"
)

const (
	FirstRunMarker   = ".gd-tools-first-run"
	HetznerTokenName = ".hetzner.token"
)

func BootstrapTest(req *Request) bool {
	path := GetRootDir(FirstRunMarker)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true
	}

	if req == nil {
		return false
	}
	return req.FQDN != "" || req.TimeZone != "" || req.SwapSize != ""
}

func BootstrapHandler(req *Request, resp *Response) error {
	if req == nil || resp == nil {
		return nil
	}

	firstRun := false
	path := GetRootDir(FirstRunMarker)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		firstRun = true
		cmds := []string{
			"apt-get update",
			"apt-get upgrade -y",
			"touch " + path,
		}
		if _, err := RunShell(cmds); err != nil {
			resp.Err = fmt.Sprintf("failed updating on first run: %v", err)
			return err
		}
		resp.Say("FirstRunMarker was created")
	} else if req.FQDN != "" {
		resp.Say("✅ FirstRunMarker exists")
	}

	// Set system hostname if provided
	if req.FQDN != "" {
		if status, err := bootstrapHostName(req.FQDN); err != nil {
			resp.Err = fmt.Sprintf("failed setting hostname to %s: %v", req.FQDN, err)
			return err
		} else {
			resp.Say(status)
		}
	}

	// Set timezone if provided
	if req.TimeZone != "" {
		if status, err := bootstrapTimeZone(req.TimeZone); err != nil {
			resp.Err = fmt.Sprintf("failed setting timezone to %s: %v", req.TimeZone, err)
			return err
		} else {
			resp.Say(status)
		}
	}

	// Create swap space if specified
	if req.SwapSize != "" && req.SwapSize != "0" {
		if status, err := bootstrapSwapSize(req.SwapSize); err != nil {
			resp.Err = fmt.Sprintf("failed setting swapsize to %s: %v", req.SwapSize, err)
			return err
		} else {
			resp.Say(status)
		}
	}

	if firstRun {
		resp.Say("New System ----- please check reboot ...")
	}

	return nil
}

// bootstrapHostName sets the system hostname using hostnamectl if needed.
func bootstrapHostName(fqdn string) (string, error) {
	currName, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("failed to read hostname: %w", err)
	}

	if currName == fqdn {
		return fmt.Sprintf("✅ hostname %s is set", fqdn), nil
	}

	if _, err := RunCommand("hostnamectl", "set-hostname", fqdn); err != nil {
		return "", fmt.Errorf("failed to set hostname: %w", err)
	}

	return fmt.Sprintf("hostname updated to: %s", fqdn), nil
}

// bootstrapTimeZone sets the system timezone using timedatectl if needed.
func bootstrapTimeZone(timeZone string) (string, error) {
	link, err := os.Readlink("/etc/localtime")
	if err != nil {
		return "", fmt.Errorf("failed to read /etc/localtime: %w", err)
	}

	if strings.HasSuffix(link, timeZone) {
		return fmt.Sprintf("✅ timezone %s is set", timeZone), nil
	}

	if _, err := RunCommand("timedatectl", "set-timezone", timeZone); err != nil {
		return "", fmt.Errorf("timedatectl failed: %w", err)
	}

	return fmt.Sprintf("timezone updated to: %s", timeZone), nil
}

// bootstrapSwapSize creates or verifies the given swap size.
func bootstrapSwapSize(swapSize string) (string, error) {
	swapFile := "/swap.img"
	if _, err := os.Stat(swapFile); err == nil {
		return fmt.Sprintf("✅ swapfile '%s' exists\n", swapFile), nil
	}

	genCmds := []string{
		fmt.Sprintf("fallocate -l %s %s", swapSize, swapFile),
		fmt.Sprintf("chmod 600 %s", swapFile),
		fmt.Sprintf("mkswap %s", swapFile),
		fmt.Sprintf("swapon %s", swapFile),
	}
	if _, err := RunShell(genCmds); err != nil {
		return "", fmt.Errorf("swapfile generation failed: %w", err)
	}

	tabCmds := []string{
		`sed -i -e '/swap/d' /etc/fstab`,
		`echo '/swap.img none swap sw 0 0' >> /etc/fstab`,
	}
	if _, err := RunShell(tabCmds); err != nil {
		return "", fmt.Errorf("failed to include swapfile in /etc/fstab: %w", err)
	}

	return fmt.Sprintf("swapfile '%s' with %s exists\n", swapFile, swapSize), nil
}
