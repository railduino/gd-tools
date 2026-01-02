package config

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/railduino/gd-tools/acme"
)

const (
	ACME_Cert_Dir    = "acme-certs"
	ACME_Account_Key = "../acme_account.key"
	CertInfoName     = "cert_info.json"
)

type CertInfo struct {
	DNS_Key    string    `json:"dns_key"`
	Domain     string    `json:"domain"`
	SANList    []string  `json:"san_list"`
	IssuedAt   time.Time `json:"issued_at"`
	ValidUntil time.Time `json:"valid_until"`
}

func (cfg *Config) EnsureCertificate(domain string, sans ...string) error {
	if domain == "" {
		return fmt.Errorf("missing domain in certificate request")
	}
	force := cfg.Force

	sans = UniqueSortedStrings(sans)
	domains := append([]string{domain}, sans...)
	baseDir := filepath.Join(ACME_Cert_Dir, domain)

	var certInfo CertInfo
	certInfoPath := filepath.Join(baseDir, CertInfoName)
	content, err := os.ReadFile(certInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			cfg.Sayf("%s not found, requesting new", certInfoPath)
			force = true
		} else {
			return fmt.Errorf("failed to read %s: %w", certInfoPath, err)
		}
	} else {
		if err := json.Unmarshal(content, &certInfo); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", certInfoPath, err)
		}
		if time.Until(certInfo.ValidUntil) <= 30*24*time.Hour {
			force = true
		}
	}

	if !force {
		fullchain, err1 := os.ReadFile(filepath.Join(baseDir, "fullchain.pem"))
		_, err2 := os.ReadFile(filepath.Join(baseDir, "privkey.pem"))
		_, err3 := os.ReadFile(filepath.Join(baseDir, "issuer.pem"))
		if err1 == nil && err2 == nil && err3 == nil {
			sanList, err := ExtractSANs(fullchain)
			if err != nil {
				return err
			}
			cfg.Sayf("✅ certificate: %s", sanList)
			cfg.PushCerts()
			return nil
		}
	}

	if cfg.Dry {
		cfg.Sayf("[dry] obtaining Certificate for %s", domain)
		return nil
	}

	key, err := acme.GetPrivateKey(ACME_Account_Key)
	if err != nil {
		return err
	}

	var resource *certificate.Resource
	provider := ""
	if cfg.HetznerToken != "" {
		provider = "Hetzner"
		os.Setenv("HETZNER_API_TOKEN", cfg.HetznerToken)
		defer os.Unsetenv("HETZNER_API_TOKEN")
		resource, err = acme.GetHetznerCertificate(domains, cfg.SysAdmin, key)
	} else if cfg.IonosToken != "" {
		provider = "IONOS"
		os.Setenv("IONOS_API_KEY", cfg.IonosToken)
		defer os.Unsetenv("IONOS_API_KEY")
		resource, err = acme.GetIonosCertificate(domains, cfg.SysAdmin, key)
	}
	// add other providers here

	if err != nil {
		return err
	}
	if resource == nil {
		return fmt.Errorf("missing provider for DNS-01 certificates")
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to mkdir %s: %w", baseDir, err)
	}

	write := func(name string, data []byte) error {
		path := filepath.Join(baseDir, name)
		return os.WriteFile(path, data, 0644)
	}

	if err := write("fullchain.pem", resource.Certificate); err != nil {
		return err
	}
	if err := write("privkey.pem", resource.PrivateKey); err != nil {
		return err
	}
	if err := write("issuer.pem", resource.IssuerCertificate); err != nil {
		return err
	}

	issuedAt := time.Now().UTC().Round(0)
	validUntil, err := ReadValidUntil(resource.Certificate)
	if err != nil {
		return err
	}

	certInfo = CertInfo{
		DNS_Key:    provider,
		Domain:     domain,
		SANList:    sans,
		IssuedAt:   issuedAt,
		ValidUntil: validUntil,
	}

	content, err = json.MarshalIndent(certInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s: %w", certInfoPath, err)
	}
	if err := os.WriteFile(certInfoPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", certInfoPath, err)
	}

	cfg.PushCerts()

	return nil
}

func ExtractSANs(fullchain []byte) (string, error) {
	var certBlock *pem.Block
	var rest = fullchain

	for {
		certBlock, rest = pem.Decode(rest)
		if certBlock == nil {
			return "", fmt.Errorf("no certificate found in fullchain")
		}
		if certBlock.Type == "CERTIFICATE" {
			break
		}
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return "", fmt.Errorf("unable to parse certificate: %w", err)
	}

	return strings.Join(cert.DNSNames, " "), nil
}

func UniqueSortedStrings(in []string) []string {
	m := make(map[string]struct{}, len(in))
	for _, s := range in {
		m[s] = struct{}{}
	}

	out := make([]string, 0, len(m))
	for s := range m {
		out = append(out, s)
	}

	sort.Strings(out)
	return out
}

func ReadValidUntil(pemBytes []byte) (time.Time, error) {
	var block *pem.Block
	rest := pemBytes

	for {
		if block, rest = pem.Decode(rest); block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err == nil {
				return cert.NotAfter.UTC().Round(0), nil
			}
		}
	}

	return time.Time{}, fmt.Errorf("no certificate in response")
}
