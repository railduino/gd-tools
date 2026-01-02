package agent

import (
	"fmt"
	"os"
	"strings"
)

const (
	FstabName = "/etc/fstab"
)

type Mount struct {
	Provider string `json:"provider"` // e.g. "hetzner"
	ID       string `json:"id"`       // e.g. "123456789"
	Dir      string `json:"dir"`      // e.g. "/var/gd-tools"
}

func (mnt *Mount) String() string {
	return fmt.Sprintf("Provider='%s' ID='%s' Dir='%s'", mnt.Provider, mnt.ID, mnt.Dir)
}

// MountsTest checks if there is work to be done
func MountsTest(req *Request) bool {
	return req != nil && len(req.Mounts) > 0
}

func MountsHandler(req *Request, resp *Response) error {
	if req == nil || resp == nil {
		return nil
	}

	for _, mount := range req.Mounts {
		if mount.ID == "" || mount.Dir == "" {
			resp.Sayf("skipped empty mount entry: %+v", mount)
			continue
		}

		switch provider := strings.ToLower(mount.Provider); provider {
		case "hetzner":
			if status, err := mountHetzner(mount.ID, mount.Dir); err != nil {
				return err
			} else {
				resp.Say(status)
			}
		case "raid":
			if status, err := mountRAID(mount.ID, mount.Dir); err != nil {
				return err
			} else {
				resp.Say(status)
			}
		default:
			return fmt.Errorf("Mount provider '%s' is not implemented", provider)
		}
	}

	return nil
}

func mountHetzner(id, dir string) (string, error) {
	if _, err := os.Stat(dir + "/lost+found"); err == nil {
		cmds := []string{
			"chown root:root " + dir,
			"chmod 0755 " + dir,
		}
		if _, err := RunShell(cmds); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ Volume %s is mounted on %s", id, dir), nil
	}

	legacy := "/mnt/HC_Volume_" + id
	if _, err := os.Stat(legacy + "/lost+found"); err == nil {
		if _, err := RunCommand("umount", legacy); err != nil {
			return "", err
		}
	} else if os.IsNotExist(err) {
		return "", fmt.Errorf("Volume %s not found: %v", legacy, err)
	}

	cmds := []string{
		"mkdir -p " + dir,
		fmt.Sprintf("sed -i -e s#%s#%s# /etc/fstab", legacy, dir),
		"systemctl daemon-reload",
		"mount -a",
		"rmdir " + legacy,
		"chown root:root " + dir,
		"chmod 0755 " + dir,
	}

	if _, err := RunShell(cmds); err != nil {
		return "", err
	}

	return fmt.Sprintf("Hetzner-Volume %s mounted on %s", id, dir), nil
}

func mountRAID(id, dir string) (string, error) {
	if _, err := os.Stat(dir + "/lost+found"); err == nil {
		return fmt.Sprintf("✅ RAID %s mounted on %s", id, dir), nil
	}

	if _, err := RunCommand("mkdir", "-p", dir); err != nil {
		return "", err
	}

	blkid, err := RunCommand("blkid", "-s", "UUID", "-o", "value", id)
	if err != nil {
		return "", err
	}
	uuid := strings.TrimSpace(string(blkid))

	data, err := os.ReadFile(FstabName)
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(string(data))
	if strings.Contains(content, uuid) {
		return "", fmt.Errorf("%s is in %s - please login and mount", id, FstabName)
	}

	line := fmt.Sprintf("/dev/disk/by-uuid/%s %s ext4 defaults,nofail 0 0", uuid, dir)
	lines := strings.Split(content, "\n")
	lines = append(lines, line)

	data = []byte(strings.Join(lines, "\n") + "\n")
	if err := os.WriteFile(FstabName, data, 0644); err != nil {
		return "", err
	}

	if _, err := RunCommand("systemctl", "daemon-reload"); err != nil {
		return "", fmt.Errorf("daemon-reload failed: %v", err)
	}

	return "", fmt.Errorf("%s has been added to %s - please login and mount", id, FstabName)
}
