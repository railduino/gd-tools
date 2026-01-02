package agent

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

func ConnectToAgent(fqdn string, seconds int, debug bool) (*tls.Conn, error) {
	crtPath := filepath.Join("CA", "client.crt")
	keyPath := filepath.Join("CA", "client.key")
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	if err != nil {
		return nil, err
	}

	caPath := filepath.Join("CA", "ca.crt")
	caCert, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCert)

	target := net.JoinHostPort(fqdn, GetAgentPort())
	if debug {
		fmt.Printf("[run] connect to: %s\n", target)
	}

	dialer := &net.Dialer{
		Timeout: time.Duration(seconds) * time.Second,
	}

	rawConn, err := dialer.Dial("tcp", target)
	if err != nil {
		return nil, err
	}

	// Wrap rawConn with TLS
	tlsConfig := &tls.Config{
		ServerName:   fqdn, // required for SNI and hostname verification
		RootCAs:      caPool,
		Certificates: []tls.Certificate{cert},
	}
	tlsConn := tls.Client(rawConn, tlsConfig)

	// Perform TLS handshake (can also time out)
	if err := tlsConn.Handshake(); err != nil {
		rawConn.Close()
		return nil, err
	}

	return tlsConn, nil
}
