package main

import (
	"crypto/rand"
	"encoding/hex"
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

	ServePort string `json:"serve_port"`

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

	if err := InitServeWeb("127.0.0.1:3000"); err != nil {
		return err
	}

	return nil
}

func NewToken(len int) (string, error) {
	token := make([]byte, len)

	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func CRC32(text string) string {
	crc := crc32.ChecksumIEEE([]byte(text))

	return fmt.Sprintf("%08X", crc)
}
