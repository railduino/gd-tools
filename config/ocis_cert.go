package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"time"
)

// EnsureIDMSelfSignedCert creates a 10-year self-signed cert for localhost/127.0.0.1/::1
// at acme-certs/ocis-idm/ldap.crt and ldap.key if they don't already exist.
func (cfg *Config) EnsureIDMSelfSignedCert() error {
	certPath := "ocis-ldap.crt"
	keyPath := "ocis-ldap.key"

	if exists(certPath) && exists(keyPath) {
		cfg.Sayf("✅ IDM self-signed certificate: %s", certPath)
		cfg.Sayf("✅ IDM self-signed private key: %s", keyPath)
		return nil
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}

	notBefore := time.Now().Add(-5 * time.Minute)
	notAfter := notBefore.Add(3650 * 24 * time.Hour) // 10 years

	tpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &priv.PublicKey, priv)
	if err != nil {
		return err
	}
	if err := writePEM(certPath, 0o644, "CERTIFICATE", der); err != nil {
		return err
	}
	cfg.Sayf("created IDM self-signed certificate %s", certPath)

	keyDER, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return err
	}
	if err := writePEM(keyPath, 0o600, "PRIVATE KEY", keyDER); err != nil {
		return err
	}
	cfg.Sayf("created IDM self-signed private key %s", keyPath)

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func writePEM(path string, mode os.FileMode, typ string, der []byte) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: typ, Bytes: der})
}
