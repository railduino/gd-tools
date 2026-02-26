package agent

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	DefaultPort     = "5320"
	ProtocolVersion = 1
)

// Request is the structure for Dev to Prod
type Request struct {
	Version int       `json:"version"`
	Conn    *tls.Conn `json:"-"`
	Verbose bool      `json:"-"`
	Quiet   bool      `json:"-"`
	Dry     bool      `json:"-"`

	Hello     string      `json:"hello,omitempty"`
	FQDN      string      `json:"fqdn,omitempty"`
	TimeZone  string      `json:"time_zone,omitempty"`
	Language  string      `json:"language,omitempty"`
	Region    string      `json:"region,omitempty"`
	SwapSize  string      `json:"swap_size,omitempty"`
	PrivRSA   []byte      `json:"priv_rsa,omitempty"`
	PublRSA   []byte      `json:"publ_rsa,omitempty"`
	Packages  []string    `json:"packages,omitempty"`
	Upgrade   bool        `json:"upgrade,omitempty"`
	SysAdmin  string      `json:"sys_admin,omitempty"`
	RedisPort int         `json:"redis_port,omitempty"`
	Mounts    []*Mount    `json:"mounts,omitempty"`
	Users     []*User     `json:"users,omitempty"`
	Files     []*File     `json:"files,omitempty"`
	Downloads []*Download `json:"downloads,omitempty"`
	SQL       []string    `json:"sql,omitempty"`
	MySQLs    []*MySQL    `json:"mysqls,omitempty"`

	Checks   []string `json:"checks,omitempty"`
	Services []string `json:"services,omitempty"` // services to (re)start
	Firewall []string `json:"firewall,omitempty"` // ports or apps to open

	// for RustDesk
	RustDesk *RustDesk `json:"rust_desk,omitempty"`

	// for Nextcloud
	Nextcloud *Nextcloud `json:"nextcloud,omitempty"`
	NextConf  string     `json:"next_conf,omitempty"`

	// for MediaWiki
	MediaWiki *MediaWiki `json:"mediawiki,omitempty"`
	MWConf    string     `json:"mw_conf,omitempty"`

	// for Ubuntu Pro (https://ubuntu.com/pro)
	UbuntuPro string `json:"ubuntu_pro,omitempty"`
}

// Response is the structure for Prod back to Dev
type Response struct {
	Result   []string  `json:"result,omitempty"`
	Services []string  `json:"services,omitempty"`
	UserIDs  []UserID  `json:"user_ids,omitempty"`
	RustDesk *RustDesk `json:"rust_desk,omitempty"`
	Err      string    `json:"err"`
}

var (
	agentPort = DefaultPort
)

func SetAgentPort(port string) {
	agentPort = port
}

func GetAgentPort() string {
	if agentPort == "" {
		return DefaultPort
	}
	return agentPort
}

func (req *Request) String() string {
	content, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return "failed to Marshal Request"
	}

	return string(content)
}

func (req *Request) AddFile(file *File) {
	if file == nil {
		return
	}
	req.Files = append(req.Files, file)
}

func (req *Request) AddService(service string) {
	if service != "" {
		req.Services = append(req.Services, service)
	}
}

func (req *Request) AddFirewall(port string) {
	if port != "" {
		req.Firewall = append(req.Firewall, port)
	}
}

func (resp *Response) String() string {
	content, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "failed to Marshal Response"
	}

	return string(content)
}

func (resp *Response) Say(args ...any) {
	msg := fmt.Sprint(args...)
	for _, line := range strings.Split(msg, "\n") {
		if line == "" {
			continue
		}
		resp.Result = append(resp.Result, line)
	}
}

func (resp *Response) Sayf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	resp.Result = append(resp.Result, msg)
}

func (resp *Response) AddService(service string) {
	if service == "" {
		return
	}

	for _, check := range resp.Services {
		if check == service {
			return
		}
	}

	resp.Services = append(resp.Services, service)
}
