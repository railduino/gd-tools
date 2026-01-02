package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/railduino/gd-tools/agent"
	"github.com/railduino/gd-tools/templates"
)

const (
	CommonNameCA     = "/CN=MyCA"
	CommonNameDev    = "/CN=dev"
	CommonNameProd   = "/CN=prod"
	SerialName       = "CA/serial"
	CaKeyName        = "CA/ca.key"
	CaCrtName        = "CA/ca.crt"
	ClientKeyName    = "CA/client.key"
	ClientCsrName    = "CA/client.csr"
	ClientCrtName    = "CA/client.crt"
	ServerKeyName    = "CA/server.key"
	ServerConfigName = "CA/server.config"
	ServerCsrName    = "CA/server.csr"
	ServerCrtName    = "CA/server.crt"
)

// AddHostToCA ensures a local Certificate Authority (CA) and generates
// a server certificate with Subject Alternative Names (SANs).
//
// The provided fqdn is added to the SAN list if not already present.
// The list is sorted to ensure idempotent certificate configuration.
func (cfg *Config) SetupCA() error {
	if err := os.MkdirAll("CA", 0700); err != nil {
		return fmt.Errorf("failed to mkdir CA: %w", err)
	}

	if ok := CheckFile(SerialName); !ok {
		if err := os.WriteFile(SerialName, []byte("1000\n"), 0600); err != nil {
			return fmt.Errorf("failed to write %s: %w", SerialName, err)
		}
	}

	if ok := CheckFile(CaKeyName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"genrsa",
			"-out", CaKeyName,
			"2048",
		); err != nil {
			return err
		}
	}

	if ok := CheckFile(CaCrtName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"req", "-x509", "-new", "-nodes",
			"-key", CaKeyName,
			"-subj", CommonNameCA,
			"-days", "3650",
			"-out", CaCrtName,
		); err != nil {
			return err
		}
	}

	if ok := CheckFile(ClientKeyName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"genrsa",
			"-out", ClientKeyName,
			"2048",
		); err != nil {
			return err
		}
	}

	if ok := CheckFile(ClientCsrName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"req", "-new",
			"-key", ClientKeyName,
			"-subj", CommonNameDev,
			"-out", ClientCsrName,
		); err != nil {
			return err
		}
	}

	if ok := CheckFile(ClientCrtName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"x509", "-req",
			"-in", ClientCsrName,
			"-CA", CaCrtName,
			"-CAkey", CaKeyName,
			"-CAserial", SerialName,
			"-out", ClientCrtName,
			"-days", "3650",
		); err != nil {
			return err
		}
	}

	if ok := CheckFile(ServerKeyName); !ok {
		if _, err := agent.RunCommand(
			"openssl",
			"genrsa",
			"-out", ServerKeyName,
			"2048",
		); err != nil {
			return err
		}
	}

	hostEntries := []string{
		"DNS.1 = " + cfg.FQDN(),
	}

	data := struct {
		HostEntries []string
	}{
		HostEntries: hostEntries,
	}
	content, err := templates.Parse("ca.server.config", cfg.Verbose, data)
	if err != nil {
		return err
	}

	existing, err := os.ReadFile(ServerConfigName)
	if err == nil && bytes.Equal(existing, content) {
		return nil
	}

	if err := os.WriteFile(ServerConfigName, content, 0600); err != nil {
		return fmt.Errorf("failed to write %s: %w", ServerConfigName, err)
	}

	if _, err := agent.RunCommand(
		"openssl",
		"req", "-new",
		"-key", ServerKeyName,
		"-subj", CommonNameProd,
		"-out", ServerCsrName,
		"-config", ServerConfigName,
	); err != nil {
		return err
	}

	if _, err := agent.RunCommand(
		"openssl",
		"x509", "-req",
		"-in", ServerCsrName,
		"-CA", CaCrtName,
		"-CAkey", CaKeyName,
		"-CAserial", SerialName,
		"-out", ServerCrtName,
		"-days", "3650",
		"-extfile", ServerConfigName,
		"-extensions", "v3_req",
	); err != nil {
		return err
	}

	return nil
}

func CheckFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 1)
	n, err := f.Read(buf)
	if err != nil || n == 0 {
		return false
	}

	return true
}
