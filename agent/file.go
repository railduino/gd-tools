package agent

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type File struct {
	Task    string `json:"task"`
	Path    string `json:"path"`
	Content []byte `json:"content,omitempty"`
	Target  string `json:"target,omitempty"`
	Backup  bool   `json:"backup,omitempty"`
	Mode    string `json:"mode,omitempty"`
	User    string `json:"user,omitempty"`
	Group   string `json:"group,omitempty"`
	Service string `json:"service,omitempty"`
}

func (f *File) String() string {
	return fmt.Sprintf("Task='%s' Path='%s'", f.Task, f.Path)
}

var (
	modeRegex = regexp.MustCompile(`^[0-7]{3,4}$`)
	RootDir   string
	VarDir    string
	ToolsDir  string
	EtcDir    string
	BinDir    string
)

func SetRootDir(path string) {
	if path == "" {
		path = os.Getenv("GD_TOOLS_ROOT_DIR")
	}
	if path == "" {
		path = "/root"
	}
	RootDir = path
}

func GetRootDir(paths ...string) string {
	SetRootDir("")
	if len(paths) == 0 {
		return RootDir
	}
	return filepath.Join(append([]string{RootDir}, paths...)...)
}

func SetVarDir(path string) {
	if path == "" {
		path = os.Getenv("GD_TOOLS_VAR_DIR")
	}
	if path == "" {
		path = "/var"
	}
	VarDir = path
}

func GetVarDir(paths ...string) string {
	SetVarDir("")
	if len(paths) == 0 {
		return VarDir
	}
	return filepath.Join(append([]string{VarDir}, paths...)...)
}

func SetToolsDir(path string) {
	if path == "" {
		path = os.Getenv("GD_TOOLS_TOOLS_DIR")
	}
	if path == "" {
		path = "/var/gd-tools"
	}
	ToolsDir = path
}

func GetToolsDir(paths ...string) string {
	SetToolsDir("")
	if len(paths) == 0 {
		return ToolsDir
	}
	return filepath.Join(append([]string{ToolsDir}, paths...)...)
}

func SieveBefore(path string) string {
	if path != "" {
		return GetToolsDir("data", "sieve_before", path)
	}
	return GetToolsDir("data", "sieve_before")
}

func SieveAfter(path string) string {
	if path != "" {
		return GetToolsDir("data", "sieve_after", path)
	}
	return GetToolsDir("data", "sieve_after")
}

func SetEtcDir(path string) {
	if path == "" {
		path = os.Getenv("GD_TOOLS_ETC_DIR")
	}
	if path == "" {
		path = "/etc"
	}
	EtcDir = path
}

func GetEtcDir(paths ...string) string {
	SetEtcDir("")
	if len(paths) == 0 {
		return EtcDir
	}
	return filepath.Join(append([]string{EtcDir}, paths...)...)
}

func SetBinDir(path string) {
	if path == "" {
		path = os.Getenv("GD_TOOLS_BIN_DIR")
	}
	if path == "" {
		path = "/usr/local/bin"
	}
	BinDir = path
}

func GetBinDir(paths ...string) string {
	SetBinDir("")
	if len(paths) == 0 {
		return BinDir
	}
	return filepath.Join(append([]string{BinDir}, paths...)...)
}

// FilesTest checks if there is work to be done
func FilesTest(req *Request) bool {
	return req != nil && len(req.Files) > 0
}

func FilesHandler(req *Request, resp *Response) error {
	if req == nil || len(req.Files) == 0 || resp == nil {
		return nil
	}

	for _, file := range req.Files {
		if file.Path == "" {
			return fmt.Errorf("missing path for %s", file.Task)
		}

		var result string
		var err error

		switch file.Task {
		case "mkdir":
			result, err = file.Mkdir(resp)
		case "write":
			result, _, err = file.Write(resp)
		case "extract":
			result, err = file.Extract(resp)
		case "read":
			content, readErr := file.Read()
			if readErr != nil {
				err = readErr
			} else {
				result = strings.Join(content, "\n")
			}
		case "delete":
			result, err = file.Delete(resp)
		case "process":
			result, err = file.Process(resp)
		case "postmap":
			result, err = file.Postmap(resp)
		case "cleanup":
			result, err = file.Cleanup(resp)
		default:
			return fmt.Errorf("unknown task '%s' for %s", file.Task, file.Path)
		}

		if err != nil {
			resp.Err = err.Error()
			return err
		}

		if result != "" {
			resp.Say(result)
		}
	}

	return nil
}

func (file *File) Mkdir(resp *Response) (string, error) {
	if stat, err := os.Stat(file.Path); err == nil && stat.IsDir() {
		if err := file.postProcess(); err != nil {
			return "", err
		}
		return fmt.Sprintf("✅ dir: %s", file.Path), nil
	}

	if err := os.MkdirAll(file.Path, 0755); err != nil {
		return "", fmt.Errorf("mkdir failed for %s: %w", file.Path, err)
	}

	if err := file.postProcess(); err != nil {
		return "", err
	}

	if file.Service != "" {
		resp.AddService(file.Service)
	}

	return fmt.Sprintf("%s successfully created", file.Path), nil
}

func (file *File) Read() ([]string, error) {
	f, err := os.Open(file.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", file.Path, err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return lines, nil
}

func (file *File) Write(resp *Response) (string, bool, error) {
	existing, err := os.ReadFile(file.Path)
	if err == nil {
		if bytes.Equal(existing, file.Content) {
			if err := file.postProcess(); err != nil {
				return "", false, err
			}
			if _, _, ok := file.IsApacheAvailable(); ok {
				return fmt.Sprintf("✅ Apache enabled: %s", file.Path), true, nil
			}
			if strings.Contains(file.Path, "systemd/system") {
				return fmt.Sprintf("✅ systemd service: %s", file.Path), true, nil
			}
			return fmt.Sprintf("✅ file: %s", file.Path), false, nil
		}
	} else if !os.IsNotExist(err) {
		return "", false, fmt.Errorf("failed to read %s: %w", file.Path, err)
	}

	if file.Backup {
		path := file.Path + ".bkup"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			os.WriteFile(path, existing, 0644)
		}
	}

	if len(file.Content) == 0 {
		f, err := os.OpenFile(file.Path, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return "", false, fmt.Errorf("failed to create %s: %w", file.Path, err)
		}
		if err := f.Close(); err != nil {
			return "", false, fmt.Errorf("failed to close %s: %w", file.Path, err)
		}
	} else {
		if err := os.WriteFile(file.Path, file.Content, 0644); err != nil {
			return "", false, fmt.Errorf("failed to write %s: %w", file.Path, err)
		}
	}

	if err := file.postProcess(); err != nil {
		return "", false, err
	}

	if file.Service != "" {
		resp.AddService(file.Service)
	}

	if _, _, ok := file.IsApacheAvailable(); ok {
		return fmt.Sprintf("✅ Apache enabled: %s", file.Path), true, nil
	}

	if strings.Contains(file.Path, "systemd/system") {
		return fmt.Sprintf("✅ systemd service: %s", file.Path), true, nil
	}

	return fmt.Sprintf("%s successfully written", file.Path), true, nil
}

func (file *File) Extract(resp *Response) (string, error) {
	if _, err := os.Stat(file.Path); err != nil {
		return "", fmt.Errorf("missing archive to extract: %s", file.Path)
	}
	if file.Target == "" {
		return "", fmt.Errorf("no target path for extract: %s", file.Path)
	}

	if err := os.MkdirAll(file.Target, 0755); err != nil {
		return "", fmt.Errorf("failed to create target %s: %w", file.Target, err)
	}

	marker := filepath.Join(file.Target, ".extracted")
	if _, err := os.Stat(marker); err == nil {
		return fmt.Sprintf("✅ extracted: %s to %s", file.Path, file.Target), nil
	}

	if err := ExtractWithCmd(file.Path, file.Target); err != nil {
		return "", fmt.Errorf("failed to extract %s to %s: %w", file.Path, file.Target, err)
	}

	if file.User != "" && file.Group != "" {
		owner_group := file.User + ":" + file.Group
		if err := ChownR(file.Target, owner_group); err != nil {
			return "", fmt.Errorf("failed to chown-R %s: %w", file.Target, err)
		}
	}

	if err := os.WriteFile(marker, []byte("ok"), 0644); err != nil {
		return "", fmt.Errorf("failed write marker for %s: %w", file.Target, err)
	}

	return fmt.Sprintf("%s extracted to %s", file.Path, file.Target), nil
}

func (file *File) Delete(resp *Response) (string, error) {
	if kind, module, ok := file.IsApacheAvailable(); ok {
		if err := file.ApacheDisable(kind, module); err != nil {
			return "", fmt.Errorf("failed to disable apache %s: %w", kind, err)
		}
	}

	if _, err := os.Stat(file.Path); err == nil {
		if err := os.Remove(file.Path); err != nil {
			return "", fmt.Errorf("failed to delete %s: %w", file.Path, err)
		}
		return fmt.Sprintf("✅ deleted: %s", file.Path), nil
	}

	return "", nil
}

func (file *File) Process(resp *Response) (string, error) {
	if err := file.postProcess(); err != nil {
		return "", err
	}

	if file.User != "" && file.Group != "" {
		owner_group := file.User + ":" + file.Group
		if err := ChownR(file.Path, owner_group); err != nil {
			return "", fmt.Errorf("failed to chown-R %s: %w", file.Target, err)
		}
	}

	if _, _, ok := file.IsApacheAvailable(); ok {
		return fmt.Sprintf("✅ Apache enabled: %s", file.Path), nil
	}

	if strings.Contains(file.Path, "systemd/system") {
		return fmt.Sprintf("✅ systemd service: %s", file.Path), nil
	}

	return fmt.Sprintf("✅ %s updated", file.Path), nil
}

func (file *File) Postmap(resp *Response) (string, error) {
	srcPath := file.Path
	dbPath := srcPath + ".db"
	tmpPath := srcPath + ".tmp"

	if err := os.WriteFile(tmpPath, file.Content, 0644); err != nil {
		return "", fmt.Errorf("failed to write %s: %w", tmpPath, err)
	}
	defer func() { _ = os.Remove(tmpPath) }()

	if err := os.Rename(tmpPath, srcPath); err != nil {
		return "", fmt.Errorf("failed to rename %s -> %s: %w", tmpPath, srcPath, err)
	}

	if _, err := RunCommand("postmap", srcPath); err != nil {
		return "", err
	}

	if err := os.Remove(srcPath); err != nil && !os.IsNotExist(err) {
		return "", fmt.Errorf("failed to remove %s: %w", srcPath, err)
	}

	oldPath := file.Path
	file.Path = dbPath
	defer func() { file.Path = oldPath }()

	if err := file.postProcess(); err != nil {
		return "", err
	}

	if file.Service != "" {
		resp.AddService(file.Service)
	}

	return fmt.Sprintf("✅ %s postmapped", file.Path), nil
}

func (file *File) Cleanup(resp *Response) (string, error) {
	if stat, err := os.Stat(file.Path); err != nil || !stat.IsDir() {
		return "", fmt.Errorf("missing dir to validate: %s", file.Path)
	}
	if file.Target == "" {
		return "", fmt.Errorf("missing pattern for validate: %s", file.Path)
	}
	validFiles := strings.Split(string(file.Content), "\n")

	dirPath := filepath.Join(file.Path, file.Target)
	dirFiles, err := filepath.Glob(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to glob %s: %w", dirPath, err)
	}

	for _, name := range dirFiles {
		keep := false
		for _, check := range validFiles {
			if strings.HasSuffix(name, check) {
				keep = true
				break
			}
		}
		if keep {
			resp.Sayf("✅ keeping %s", name)
			continue
		}
		if err := os.Remove(name); err != nil {
			return "", fmt.Errorf("failed to delete %s: %w", name, err)
		}
		resp.Sayf("✅ deleted %s", name)
	}

	return "", nil
}

func (file *File) postProcess() error {
	if err := ApplyPermissions(file.Path, file.Mode, file.User, file.Group); err != nil {
		return err
	}

	if kind, module, ok := file.IsApacheAvailable(); ok {
		if err := file.ApacheEnable(kind, module); err != nil {
			return err
		}
	}

	if strings.Contains(file.Path, "systemd/system") {
		if err := ReloadSystemd(filepath.Base(file.Path)); err != nil {
			return err
		}
	}

	return nil
}

func ApplyPermissions(path, mode, owner, group string) error {
	if mode != "" {
		if valid := modeRegex.MatchString(mode); !valid {
			return fmt.Errorf("invalid mode '%s' for %s", mode, path)
		}
		modeValue, err := strconv.ParseUint(mode, 8, 32)
		if err != nil {
			return fmt.Errorf("failed to parse mode '%s' for %s", mode, path)
		}
		if err := os.Chmod(path, os.FileMode(modeValue)); err != nil {
			return fmt.Errorf("chmod failed: %w", err)
		}
	}

	if owner != "" || group != "" {
		usr, err := user.Lookup(owner)
		if err != nil {
			return fmt.Errorf("owner lookup failed for %s: %w", path, err)
		}
		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return fmt.Errorf("invalid UID for owner %s: %w", owner, err)
		}

		gid := -1
		if group != "" {
			grp, err := user.LookupGroup(group)
			if err != nil {
				return fmt.Errorf("group lookup failed for %s: %w", path, err)
			}
			gid, err = strconv.Atoi(grp.Gid)
			if err != nil {
				return fmt.Errorf("invalid GID for group %s: %w", group, err)
			}
		} else {
			gid, err = strconv.Atoi(usr.Gid)
			if err != nil {
				return fmt.Errorf("invalid GID for group %s: %w", group, err)
			}
		}

		if err := os.Chown(path, uid, gid); err != nil {
			return fmt.Errorf("chown failed for %s: %w", path, err)
		}
	}

	return nil
}

func ExtractWithCmd(srcFile, destDir string) error {
	var cmd *exec.Cmd

	switch {
	case strings.HasSuffix(srcFile, ".zip"):
		cmd = exec.Command("unzip", "-q", srcFile, "-d", destDir)
	case strings.HasSuffix(srcFile, ".tar.gz"), strings.HasSuffix(srcFile, ".tgz"):
		cmd = exec.Command("tar", "-xzf", srcFile, "-C", destDir)
	case strings.HasSuffix(srcFile, ".tar.bz2"):
		cmd = exec.Command("tar", "-xjf", srcFile, "-C", destDir)
	case strings.HasSuffix(srcFile, ".tar.xz"):
		cmd = exec.Command("tar", "-xJf", srcFile, "-C", destDir)
	default:
		return fmt.Errorf("unsupported archive format: %s", srcFile)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("extract failed for %s: %w", srcFile, err)
	}

	return nil
}

func ChownR(path, owner_group string) error {
	cmd := exec.Command("chown", "-R", owner_group, path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ReloadSystemd(service string) error {
	if err := exec.Command("systemctl", "daemon-reexec").Run(); err != nil {
		return fmt.Errorf("failed to daemon-reexec: %w", err)
	}
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to daemon-reload: %w", err)
	}
	if err := exec.Command("systemctl", "enable", "--now", service).Run(); err != nil {
		return fmt.Errorf("failed to enable-now %s: %w", service, err)
	}

	return nil
}
