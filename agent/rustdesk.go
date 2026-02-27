package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	RustDeskName   = "rustdesk.json"
	RustDeskNumber = 41
)

// RustDesk is the request/response payload section for managing hbbs/hbbr server keys.
// - If PrivateKeyB64 is present in the request, prod restores the keys.
// - If PrivateKeyB64 is empty and prod has keys, prod exports them in the response.
// - If keys do not exist on prod, prod boots hbbs once (or starts service) to generate them, then exports.
// - This structure is also used to save the configuration on the development system.
type RustDesk struct {
	// The host name when sent to the production server
	HostName   string `json:"host_name,omitempty"`
	DomainName string `json:"domain_name,omitempty"`

	// Keys are stored as raw file bytes encoded in base64.
	// Private key must never be logged.
	PrivateKeyB64 string `json:"private_key_b64,omitempty"`
	PublicKeyB64  string `json:"public_key_b64,omitempty"`

	// Optional metadata for safety/diagnostics.
	// Fingerprint should be derived from the public key (e.g. sha256) and is safe to log/display.
	Fingerprint string `json:"fingerprint,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"` // RFC3339, optional

	// Safety switch: allow replacing existing keys on prod (default false).
	ForceReplace bool `json:"force_replace,omitempty"`
}

func (rd *RustDesk) FQDN() string {
	return rd.HostName + "." + rd.DomainName
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

	return nil
}
