package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const (
	ServeConfigName = "gd-tools-serve.conf"
)

type ServeConfig struct {
	AdminUser string `json:"admin_user"`
	AdminMail string `json:"admin_mail"`
	AdminPswd string `json:"admin_pswd"`

	Address   string `json:"address"`
	JWTSecret string `json:"jwt_secret"`

	ImprintURL string `json:"imprint_url"`
	ProtectURL string `json:"protect_url"`
}

func init() {
	AddSubCommand(commandServe, "any") // TODO "prod"
}

var commandServe = &cli.Command{
	Name:        "serve",
	Usage:       T("serve-cmd-usage"),
	Description: T("serve-cmd-describe"),
	Flags:       []cli.Flag{},
	Action:      runServe,
}

func runServe(c *cli.Context) error {
	if euid := os.Geteuid(); euid == 0 {
		msg := T("serve-not-as-root")
		return fmt.Errorf(msg)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	if err := os.Chdir(homeDir); err != nil {
		return err
	}

	serveConfig, err := GetServeConfig()
	if err != nil {
		return err
	}

	LocaleInit()
	for _, line := range LocaleGetInfo() {
		log.Printf("Locale: %s", line)
	}

	if err := InitServeDB(); err != nil {
		return err
	}

	if err := InitServeUser(); err != nil {
		return err
	}
	UserLoginInit()

	if err := InitServeWeb(serveConfig.Address); err != nil {
		return err
	}

	return nil
}

func GetServeConfig() (*ServeConfig, error) {
	var serveConfig ServeConfig

	content, err := os.ReadFile(ServeConfigName)
	if err != nil {
		if os.IsNotExist(err) {
			serveConfig = ServeConfig{
				Address: "127.0.0.1:3000",
			}
			if err := serveConfig.Save(); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(content, &serveConfig); err != nil {
			return nil, err
		}
	}

	if serveConfig.AdminPswd == "" {
		msg := T("serve-first-edit-config")
		return nil, fmt.Errorf(msg)
	}

	return &serveConfig, nil
}

func (sc *ServeConfig) Save() error {
	content, err := json.MarshalIndent(*sc, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(ServeConfigName, content, 0644); err != nil {
		return err
	}

	return nil
}

func NewToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func CRC32(text string) string {
	crc := crc32.ChecksumIEEE([]byte(text))

	return fmt.Sprintf("%08X", crc)
}
