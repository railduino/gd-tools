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

var systemFlagProgress = cli.BoolFlag{
	Name:    "progress",
	Aliases: []string{"p"},
	Usage:   T("system-flag-progress"),
}

var systemFlagUpgrade = cli.BoolFlag{
	Name:    "upgrade",
	Aliases: []string{"u"},
	Usage:   T("system-flag-upgrade"),
}

var commandSystem = &cli.Command{
	Name:        "system",
	Usage:       T("system-cmd-usage"),
	Description: T("system-cmd-describe"),
	Flags: []cli.Flag{
		&mainFlagDryRun,
		&systemFlagProgress,
		&systemFlagUpgrade,
	},
	Action: runSystem,
}

func runSystem(c *cli.Context) error {
	localPath, err := os.Getwd()
	if err != nil {
		return err
	}
	dryRun := c.Bool("dry")

	if CheckEnv("dev") {
		hostName := filepath.Base(localPath)
		rootUser := fmt.Sprintf("root@%s", hostName)
		DeployFetchLetsEncrypt(c, rootUser)

		return ShellEditor(SystemConfigName)
	}

	// this must be "prod" - only root is allowed to run
	if euid := os.Geteuid(); euid != 0 {
		msg := T("system-only-root")
		return fmt.Errorf(msg)
	}

	systemConfigFile := filepath.Join("/etc", SystemConfigName)
	content, err := os.ReadFile(systemConfigFile)
	if err != nil {
		return err
	}

	var systemConfig SystemConfig
	if err := json.Unmarshal(content, &systemConfig); err != nil {
		return err
	}
	systemConfig.DryRun = dryRun
	systemConfig.Progress = c.Bool("progress")
	systemConfig.Upgrade = c.Bool("upgrade")

	if err := systemConfig.SetTimeZone(); err != nil {
		return err
	}
	if err := systemConfig.SetHostName(); err != nil {
		return err
	}
	if err := systemConfig.AddSwapSpace(); err != nil {
		return err
	}
	if err := systemConfig.AddDockerRepo(); err != nil {
		return err
	}
	if err := systemConfig.InstallPackages(); err != nil {
		return err
	}
	if err := systemConfig.SetupMounts(); err != nil {
		return err
	}
	if err := systemConfig.ActivateFirewall(); err != nil {
		return err
	}
	if err := systemConfig.AddToolsUser(); err != nil {
		return err
	}
	if err := systemConfig.CollectData(); err != nil {
		return err
	}

	return nil
}

func (sc *SystemConfig) SetTimeZone() error {
	currZone, err := FileGetLine("/etc/timezone")
	if err != nil {
		return err
	}
	if currZone == sc.TimeZone {
		msg := Tf("system-timezone-okay", sc.TimeZone)
		fmt.Println(msg)
		return nil
	}

	msg := Tf("system-timezone-update", sc.TimeZone)
	fmt.Println(msg)

	set_timezone := fmt.Sprintf("timedatectl set-timezone %s", sc.TimeZone)
	return ShellCmd(sc.DryRun, set_timezone)
}

func (sc *SystemConfig) SetHostName() error {
	currName, err := FileGetLine("/etc/hostname")
	if err != nil {
		return err
	}
	if currName == sc.HostName {
		msg := Tf("system-hostname-okay", sc.HostName)
		fmt.Println(msg)
		return nil
	}

	msg := Tf("system-hostname-update", sc.HostName)
	fmt.Println(msg)

	set_hostname := fmt.Sprintf("hostnamectl set-hostname %s", sc.HostName)
	return ShellCmd(sc.DryRun, set_hostname)
}

func (sc *SystemConfig) AddSwapSpace() error {
	if sc.SwapSpace <= 0 {
		fmt.Println(T("system-swapfile-zero"))
		return nil
	}
	swapSize := fmt.Sprintf("%dG", sc.SwapSpace)

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

	if err := ShellCmds(sc.DryRun, cmds); err != nil {
		return err
	}

	if sc.DryRun {
		msg := Tf("system-swapfile-fstab", swapFile)
		fmt.Println(msg)
		return nil
	}

	return FileAddLine("/etc/fstab",
		`^/swap\.img\s+none\s+swap\s+sw\s+0\s+0$`,
		"/swap.img none swap sw 0 0")
}

func (sc *SystemConfig) AddDockerRepo() error {
	dockerURL := "https://download.docker.com/linux/ubuntu"
	gpgKey := "/etc/apt/keyrings/docker.gpg"
	dockerDeb := "/etc/apt/sources.list.d/docker.list"

	if sc.DryRun {
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

func (sc *SystemConfig) InstallPackages() error {
	if err := ShellCmd(sc.DryRun, "apt update"); err != nil {
		return err
	}

	for _, pkg_name := range sc.Packages {
		pkg_info := exec.Command("dpkg", "-s", pkg_name)
		if err := pkg_info.Run(); err == nil {
			fmt.Printf("- %s ist bereits installiert\n", pkg_name)
			continue
		}

		apt_get_install := fmt.Sprintf("apt install -y %s", pkg_name)
		if err := ShellCmd(sc.DryRun, apt_get_install); err != nil {
			return err
		}
	}

	if err := SystemService(sc.DryRun, "ssh"); err != nil {
		return err
	}
	if err := SystemService(sc.DryRun, "docker"); err != nil {
		return err
	}
	if err := SystemService(sc.DryRun, "nginx"); err != nil {
		return err
	}

	return nil
}

func (sc *SystemConfig) SetupMounts() error {
	if err := ShellCmd(sc.DryRun, "mount -a"); err != nil {
		return err
	}

	for _, mount := range sc.Mounts {
		var err error
		switch provider := strings.ToLower(mount.Provider); provider {
		case "hetzner":
			err = systemMountHetzner(sc.DryRun, mount.Identifier, mount.Mountpoint)
		case "raid":
			err = systemMountRAID(sc.DryRun, mount.Identifier, mount.Mountpoint)
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

func (sc *SystemConfig) ActivateFirewall() error {
	if err := ShellCmd(sc.DryRun, "ufw allow OpenSSH"); err != nil {
		return err
	}
	if err := ShellCmd(sc.DryRun, "ufw allow Nginx_#_Full"); err != nil {
		return err
	}

	matched, err := ShellMatch("ufw status", "Status: active")
	if err != nil {
		return err
	}
	if !matched {
		if err := ShellCmd(sc.DryRun, "ufw enable"); err != nil {
			return err
		}
	}

	return nil
}

func (sc *SystemConfig) AddToolsUser() error {
	envFile := "/etc/gd-tools-env"
	if err := os.WriteFile(envFile, []byte("prod\n"), 0o444); err != nil {
		return err
	}

	gdUser, err := user.Lookup("gd-tools")
	if err != nil {
		userAdd := fmt.Sprintf("useradd -r -m -s /bin/bash gd-tools")
		if err := ShellCmd(sc.DryRun, userAdd); err != nil {
			return err
		}
	}
	gdUser, err = user.Lookup("gd-tools")
	if err != nil {
		return err
	}
	userChmod := fmt.Sprintf("chmod 755 /home/gd-tools")
	if err := ShellCmd(sc.DryRun, userChmod); err != nil {
		return err
	}

	gdGroups, err := gdUser.GroupIds()
	if err != nil {
		return err
	}
	if ok := slices.Contains(gdGroups, "docker"); !ok {
		groupAdd := fmt.Sprintf("usermod -aG docker gd-tools")
		if err := ShellCmd(sc.DryRun, groupAdd); err != nil {
			return err
		}
	}

	sshCmds := []string{
		"install -o gd-tools -g gd-tools -m 700 -d /home/gd-tools/.ssh",
		"install -o gd-tools -g gd-tools -m 600 /root/.ssh/authorized_keys /home/gd-tools/.ssh",
		"install -o gd-tools -g gd-tools -m 755 -d " + SystemDataRoot,
		"install -o gd-tools -g gd-tools -m 755 -d " + SystemLogsRoot,
	}
	if err := ShellCmds(sc.DryRun, sshCmds); err != nil {
		return err
	}

	return nil
}

func (sc *SystemConfig) CollectData() error {
	if _, err := os.ReadDir("/etc/letsencrypt"); err != nil {
		certbotOpts := fmt.Sprintf("--nginx --non-interactive --agree-tos --email %s", sc.SysAdmin)
		certbotCmd := fmt.Sprintf("certbot certonly %s -d %s", certbotOpts, sc.HostName)
		if err := ShellCmd(sc.DryRun, certbotCmd); err != nil {
			return err
		}
	}
	if _, err := os.ReadDir("/etc/letsencrypt"); err != nil {
		return err
	}

	gdtUser, err := user.Lookup("gd-tools")
	if err != nil {
		return err
	}
	dckGroup, err := user.LookupGroup("docker")
	if err != nil {
		return err
	}
	if sc.DryRun {
		msg := Tf("system-list_ids", gdtUser.Uid, gdtUser.Gid, dckGroup.Gid)
		fmt.Println("[dry]", msg)
	} else {
		uidData := SystemIDs{
			ToolsUID:  gdtUser.Uid,
			ToolsGID:  gdtUser.Gid,
			DockerGID: dckGroup.Gid,
		}
		content, err := json.MarshalIndent(uidData, "", "  ")
		if err != nil {
			return err
		}
		uidPath := filepath.Join("/etc/letsencrypt", SystemIDsName)
		if err := os.WriteFile(uidPath, content, 0644); err != nil {
			return err
		}
	}

	return nil
}
