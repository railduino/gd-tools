package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slices"
)

func init() {
	AddSubCommand(commandSystem, "any")
}

var commandSystem = &cli.Command{
	Name:        "system",
	Usage:       T("system-cmd-usage"),
	Description: T("system-cmd-describe"),
	Flags: []cli.Flag{
		&mainFlagDryRun,
	},
	Action: runSystem,
}

func runSystem(c *cli.Context) error {
	if CheckEnv("dev") {
		return ShellEditor(SystemConfigFile)
	}

	// must be "prod" - only root is allowed to run
	if euid := os.Geteuid(); euid != 0 {
		msg := T("system-only-root")
		return fmt.Errorf(msg)
	}

	systemFile := filepath.Join("/etc", SystemConfigFile)
	systemData, err := os.ReadFile(systemFile)
	if err != nil {
		return err
	}
	var systemConfig SystemConfig
	if err := json.Unmarshal(systemData, &systemConfig); err != nil {
		return err
	}

	dryRun := c.Bool("dry")
	if err := systemTimeZone(dryRun, systemConfig.TimeZone); err != nil {
		return err
	}
	if err := systemSwapFile(dryRun, systemConfig.SwapSpace); err != nil {
		return err
	}
	if err := systemHostName(dryRun, systemConfig.HostName); err != nil {
		return err
	}
	if err := systemDocker(dryRun); err != nil {
		return err
	}
	if err := systemPackages(dryRun, systemConfig.Packages); err != nil {
		return err
	}
	if err := systemMounts(dryRun, systemConfig.Mounts); err != nil {
		return err
	}
	if err := systemFirewall(dryRun, systemConfig.SshPort); err != nil {
		return err
	}
	if err := systemUser(dryRun); err != nil {
		return err
	}

	return nil
}

func systemTimeZone(dryRun bool, zone string) error {
	currZone, err := FileGetLine("/etc/timezone")
	if err != nil {
		return err
	}
	if currZone == zone {
		msg := Tf("system-timezone-okay", zone)
		fmt.Println(msg)
		return nil
	}

	msg := Tf("system-timezone-update", zone)
	fmt.Println(msg)

	set_timezone := fmt.Sprintf("timedatectl set-timezone %s", zone)
	return ShellCmd(dryRun, set_timezone)
}

func systemSwapFile(dryRun bool, size int) error {
	if size <= 0 {
		fmt.Println(T("system-swapfile-zero"))
		return nil
	}
	swapSize := fmt.Sprintf("%dG", size)

	swapFile := "/swap.img"
	if _, err := os.Stat(swapFile); err == nil {
		msg := Tf("system-swapfile-exist", swapFile)
		fmt.Println(msg)
		return nil
	}

	cmds := []string{
		fmt.Sprintf("fallocate -l %s %s", swapSize, swapFile),
		fmt.Sprintf("chmod 600 %s", swapFile),
		fmt.Sprintf("mkswap %s", swapFile),
		fmt.Sprintf("swapon %s", swapFile),
	}

	if err := ShellCmds(dryRun, cmds); err != nil {
		return err
	}

	if dryRun {
		msg := Tf("system-swapfile-fstab", swapFile)
		fmt.Println(msg)
		return nil
	}

	return FileAddLine("/etc/fstab", `^/swap\.img\s+none\s+swap\s+sw\s+0\s+0$`, "/swap.img none swap sw 0 0")
}

func systemHostName(dryRun bool, name string) error {
	currName, err := FileGetLine("/etc/hostname")
	if err != nil {
		return err
	}
	if currName == name {
		msg := Tf("system-hostname-okay", name)
		fmt.Println(msg)
		return nil
	}

	msg := Tf("system-hostname-update", name)
	fmt.Println(msg)

	set_hostname := fmt.Sprintf("hostnamectl set-hostname %s", name)
	return ShellCmd(dryRun, set_hostname)
}

func systemDocker(dryRun bool) error {
	dockerURL := "https://download.docker.com/linux/ubuntu"
	gpgKey := "/etc/apt/keyrings/docker.gpg"
	dockerDeb := "/etc/apt/sources.list.d/docker.list"

	if dryRun {
		cmd := fmt.Sprintf("install Docker from %s ...", dockerURL)
		return ShellCmd(true, cmd)
	}

	if err := os.MkdirAll("/etc/apt/keyrings", 0755); err != nil {
		return err
	}

	resp, err := http.Get(dockerURL + "/gpg")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	var out bytes.Buffer
	cmd := exec.Command("gpg", "--dearmor")
	cmd.Stdin = resp.Body
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.WriteFile(gpgKey, out.Bytes(), 0644); err != nil {
		return err
	}

	envMap, err := godotenv.Read("/etc/os-release")
	if err != nil {
		return err
	}
	codeName := envMap["VERSION_CODENAME"]

	var arch string
	switch runtime.GOARCH {
	case "amd64":
		arch = "amd64"
	case "arm64":
		arch = "arm64"
	default:
		return fmt.Errorf("unsupported GOARCH: %s", runtime.GOARCH)
	}

	aptSource := fmt.Sprintf("deb [arch=%s signed-by=%s] %s %s stable\n",
		arch, gpgKey, dockerURL, codeName)

	if err := os.WriteFile(dockerDeb, []byte(aptSource), 0644); err != nil {
		return err
	}

	return nil
}

func systemPackages(dryRun bool, packages []string) error {
	if err := ShellCmd(dryRun, "apt update"); err != nil {
		return err
	}

	for _, pkg_name := range packages {
		pkg_info := exec.Command("dpkg", "-s", pkg_name)
		if err := pkg_info.Run(); err == nil {
			fmt.Printf("- %s ist bereits installiert\n", pkg_name)
			continue
		}

		apt_get_install := fmt.Sprintf("apt install -y %s", pkg_name)
		if err := ShellCmd(dryRun, apt_get_install); err != nil {
			return err
		}
	}

	if err := SystemService(dryRun, "ssh"); err != nil {
		return err
	}
	if err := SystemService(dryRun, "docker"); err != nil {
		return err
	}

	return nil
}

func systemMounts(dryRun bool, mounts []Mount) error {
	if err := ShellCmd(dryRun, "mount -a"); err != nil {
		return err
	}

	for _, mount := range mounts {
		var err error
		switch provider := strings.ToLower(mount.Provider); provider {
		case "hetzner":
			err = systemMountHetzner(dryRun, mount.Identifier, mount.Mountpoint)
		case "raid":
			err = systemMountRAID(dryRun, mount.Identifier, mount.Mountpoint)
		default:
			err = fmt.Errorf("Provider '%s' ist noch nicht implementiert.", provider)
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func systemMountHetzner(dryRun bool, id, target string) error {
	if _, err := os.Stat(target + "/lost+found"); err == nil {
		fmt.Printf("Volume '%s' bereits eingebunden unter %s\n", id, target)
		return nil
	}

	legacy := "/mnt/HC_Volume_" + id
	if _, err := os.Stat(legacy + "/lost+found"); err == nil {
		if err := ShellCmd(dryRun, "umount "+legacy); err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		return fmt.Errorf("Volume %s nicht gefunden: %v", legacy, err)
	}

	cmds := []string{
		"mkdir -p " + target,
		fmt.Sprintf("sed -i -e s#%s#%s# /etc/fstab", legacy, target),
		"systemctl daemon-reload",
		"mount -a",
		"rmdir " + legacy,
		"chmod 0755 " + target,
	}

	return ShellCmds(dryRun, cmds)
}

func systemMountRAID(dryRun bool, id, target string) error {
	if _, err := os.Stat(target + "/lost+found"); err == nil {
		fmt.Println("RAID-Volume bereits eingebunden unter:", target)
		return nil
	}

	mkdir := fmt.Sprintf("mkdir -p %s", target)
	if err := ShellCmd(dryRun, mkdir); err != nil {
		return err
	}

	uuid, err := ShellGetDeviceUUID(dryRun, id)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("UUID=%s %s ext4 defaults,nofail 0 0", uuid, target)
	if dryRun {
		fmt.Printf("[dry] /etc/fstab: '%s'\n", line)
	} else {
		if err := FileAddLine("/etc/fstab", uuid, line); err != nil {
			return err
		}
	}

	cmds := []string{
		"systemctl daemon-reload",
		"mount -a",
	}

	return ShellCmds(dryRun, cmds)
}

func systemFirewall(dryRun bool, sshPort string) error {
	allowSSH := fmt.Sprintf("ufw allow %s", sshPort)
	if err := ShellCmd(dryRun, allowSSH); err != nil {
		return err
	}

	matched, err := ShellMatch("ufw status", "Status: active")
	if err != nil {
		return err
	}
	if !matched {
		if err := ShellCmd(dryRun, "ufw enable"); err != nil {
			return err
		}
	}

	return nil
}

func systemUser(dryRun bool) error {
	envFile := "/etc/gd-tools-env"
	if err := os.WriteFile(envFile, []byte("prod\n"), 0o444); err != nil {
		return err
	}

	gdUser, err := user.Lookup("gd-tools")
	if err != nil {
		userAdd := fmt.Sprintf("useradd -r -m -s /bin/bash gd-tools")
		if err := ShellCmd(dryRun, userAdd); err != nil {
			return err
		}
	}
	gdUser, err = user.Lookup("gd-tools")
	if err != nil {
		return err
	}

	gdGroups, err := gdUser.GroupIds()
	if err != nil {
		return err
	}
	if ok := slices.Contains(gdGroups, "docker"); !ok {
		groupAdd := fmt.Sprintf("usermod -aG docker gd-tools")
		if err := ShellCmd(dryRun, groupAdd); err != nil {
			return err
		}
	}

	dataRoot, _ := GetDataRoot("prod", "volumes")
	logsRoot, _ := GetDataRoot("prod", "logs")
	sshCmds := []string{
		"install -o gd-tools -g gd-tools -m 700 -d /home/gd-tools/.ssh",
		"install -o gd-tools -g gd-tools -m 600 /root/.ssh/authorized_keys /home/gd-tools/.ssh",
		"install -o gd-tools -g gd-tools -m 755 -d " + dataRoot,
		"install -o gd-tools -g gd-tools -m 755 -d " + logsRoot,
	}
	if err := ShellCmds(dryRun, sshCmds); err != nil {
		return err
	}

	return nil
}
