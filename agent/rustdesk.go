package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"encoding/base64"
)

const (
	RustDeskName   = "rustdesk.json"
	RustDeskNumber = 41
)

type RustDesk struct {
	HostName   string `json:"host_name,omitempty"`
	DomainName string `json:"domain_name,omitempty"`

	PrivateB64 string `json:"private_b64,omitempty"`
	PublicKey  string `json:"public_key,omitempty"`
}

func (rd *RustDesk) FQDN() string {
	return rd.HostName + "." + rd.DomainName
}

func (rd *RustDesk) GetPrivate() ([]byte, error) {
	if rd.PrivateB64 == "" {
		return nil, nil
	}
	return base64.StdEncoding.DecodeString(rd.PrivateB64)
}

func (rd *RustDesk) SetPrivate(data []byte) {
	if len(data) == 0 {
		rd.PrivateB64 = ""
		return
	}
	rd.PrivateB64 = base64.StdEncoding.EncodeToString(data)
}

func (rd *RustDesk) GetPublic() string {
	if rd.PublicKey == "" {
		return ""
	}
	return strings.TrimSpace(rd.PublicKey)
}

func (rd *RustDesk) SetPublic(s string) {
	rd.PublicKey = strings.TrimSpace(s)
}

// This is for mount checking
func (rd *RustDesk) ToolsDir() string {
	return GetToolsDir()
}

func (rd *RustDesk) DataDir(paths ...string) string {
	root := GetToolsDir("data", "rustdesk")
	if len(paths) == 0 {
		return root
	}
	return filepath.Join(append([]string{root}, paths...)...)
}

func (rd *RustDesk) LogsDir(paths ...string) string {
	root := GetToolsDir("logs", "rustdesk")
	if len(paths) == 0 {
		return root
	}
	return filepath.Join(append([]string{root}, paths...)...)
}

// The following functions are used on the development system
func (rd *RustDesk) Save() error {
	content, err := json.MarshalIndent(rd, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", RustDeskName, err)
	}

	existing, err := os.ReadFile(RustDeskName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(RustDeskName, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", RustDeskName, err)
	}

	return nil
}

// The following functions are used on the production host
func RustDeskTest(req *Request) bool {
	return req != nil && req.RustDesk != nil
}

func RustDeskHandler(req *Request, resp *Response) error {
	if RustDeskTest(req) == false {
		return nil
	}
	rd := req.RustDesk

	privPath := rd.DataDir("id_ed25519")
	privKey, err := os.ReadFile(privPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", privPath, err)
	}
	privB64 := base64.StdEncoding.EncodeToString(privKey)
	if rd.PrivateB64 != "" && privB64 != "" && rd.PrivateB64 != privB64 {
		return fmt.Errorf("RustDesk private key differs")
	}
	rd.SetPrivate(privKey)

	pubPath := rd.DataDir("id_ed25519.pub")
	pubBytes, err := os.ReadFile(pubPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", pubPath, err)
	}
	pubKey := strings.TrimSpace(string(pubBytes))
	if rd.GetPublic() != "" && pubKey != "" && rd.GetPublic() != pubKey {
		return fmt.Errorf("RustDesk public key differs")
	}
	rd.SetPublic(pubKey)

	resp.RustDesk = rd
	resp.Say("✅ RustDesk keys match")

	return nil
}
