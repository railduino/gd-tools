package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"syscall"

	"github.com/railduino/gd-tools/agent"
)

var (
	lockFile *os.File
)

type Handler struct {
	Name string
	Test func(*agent.Request) bool
	Func func(*agent.Request, *agent.Response) error
}

// N.B. order is important here (don't change the first block)
var Handlers = []Handler{
	{
		Name: "Hello",
		Test: agent.HelloTest,
		Func: agent.HelloHandler,
	},
	{
		Name: "Bootstrap",
		Test: agent.BootstrapTest,
		Func: agent.BootstrapHandler,
	},
	{
		Name: "Downloads",
		Test: agent.DownloadsTest,
		Func: agent.DownloadsHandler,
	},
	{
		Name: "Packages",
		Test: agent.PackagesTest,
		Func: agent.PackagesHandler,
	},
	{
		Name: "Mounts",
		Test: agent.MountsTest,
		Func: agent.MountsHandler,
	},
	{
		Name: "Redis",
		Test: agent.RedisTest,
		Func: agent.RedisHandler,
	},
	{
		Name: "Users",
		Test: agent.UsersTest,
		Func: agent.UsersHandler,
	},
	{
		Name: "Files",
		Test: agent.FilesTest,
		Func: agent.FilesHandler,
	},
	{
		Name: "MySQL",
		Test: agent.MySQLTest,
		Func: agent.MySQLHandler,
	},
	{
		Name: "MediaWiki",
		Test: agent.MediaWikiTest,
		Func: agent.MediaWikiHandler,
	},
	{
		Name: "Nextcloud",
		Test: agent.NextcloudTest,
		Func: agent.NextcloudHandler,
	},
	{
		Name: "OCIS",
		Test: agent.OCISTest,
		Func: agent.OCISHandler,
	},
	{
		Name: "RustDesk",
		Test: agent.RustDeskTest,
		Func: agent.RustDeskHandler,
	},
}

func acquireAgentLock() {
	var err error

	lockFile, err = os.OpenFile("/var/lock/gd-tools-agent.lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		log.Fatalf("failed to open lock file: %v", err)
	}

	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		log.Fatalf("another instance of gd-tools-agent is already running")
	}
}

func main() {
	acquireAgentLock()
	defer func() {
		if lockFile != nil {
			syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN)
			lockFile.Close()
		}
	}()

	port := flag.String("port", agent.DefaultPort, "Port to listen on (default 5320)")
	verbose := flag.Bool("verbose", false, "Enable debug output")
	flag.Parse()

	tlsCert, err := tls.LoadX509KeyPair(
		agent.GetEtcDir("gd-tools", "server.crt"),
		agent.GetEtcDir("gd-tools", "server.key"),
	)
	if err != nil {
		log.Fatalf("Failed to load server certificate: %v", err)
	}

	caCertBytes, err := os.ReadFile(
		agent.GetEtcDir("gd-tools", "ca.crt"),
	)
	if err != nil {
		log.Fatalf("Failed to load CA certificate: %v", err)
	}

	caPool := x509.NewCertPool()
	caPool.AppendCertsFromPEM(caCertBytes)

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
	}

	if *verbose {
		block, _ := pem.Decode(caCertBytes)
		if block == nil {
			log.Fatalf("CA fingerprint error: invalid PEM data")
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Fatalf("CA fingerprint parse error: %v", err)
		}
		sum := sha256.Sum256(cert.Raw)
		log.Printf("CA SHA256 Fingerprint: %s", formatFingerprint(sum))
	}

	address := ":" + *port
	ln, err := tls.Listen("tcp", address, tlsCfg)
	if err != nil {
		log.Fatalf("failed to listen on port %s: %v", *port, err)
	}
	log.Printf("INFO: gd-tools agent started on port %s", *port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept failed: %v", err)
			continue
		}

		go handleConnection(conn)
	}
}

func formatFingerprint(sum [32]byte) string {
	var parts []string
	for _, b := range sum {
		parts = append(parts, fmt.Sprintf("%02X", b))
	}
	return strings.Join(parts, ":")
}

func handleConnection(c net.Conn) {
	defer c.Close()

	for {
		var req agent.Request
		var resp agent.Response

		decoder := json.NewDecoder(c)
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				// client closed connection, no real error
				return
			}
			err := fmt.Errorf("failed to decode command: %w", err)
			EncodeResponse(c, &resp, err)
			return
		}

		if req.Version != agent.ProtocolVersion {
			err := fmt.Errorf("protocol mismatch, expected %d, got %d", agent.ProtocolVersion, req.Version)
			EncodeResponse(c, &resp, err)
			return
		}

		for _, service := range req.Services {
			resp.AddService(service)
		}

		for _, handler := range Handlers {
			if handler.Test(&req) {
				log.Printf("entering handler: %s", handler.Name)
				if err := handler.Func(&req, &resp); err != nil {
					err := fmt.Errorf("failed to complete handler %s: %v", handler.Name, err)
					EncodeResponse(c, &resp, err)
					return
				}
				log.Printf("leaving handler: %s", handler.Name)
			}
		}

		for _, service := range resp.Services {
			status, err := agent.StartService(service)
			if err != nil {
				EncodeResponse(c, &resp, err)
				return
			}
			resp.Say(status)
		}

		if len(req.Firewall) > 0 {
			if err := resp.FirewallOpen(c, req.Firewall); err != nil {
				EncodeResponse(c, &resp, err)
				return
			}
		}

		EncodeResponse(c, &resp, nil)
	}
}

func EncodeResponse(c net.Conn, resp *agent.Response, err error) {
	if err != nil {
		log.Printf("%s", err.Error())
		resp.Err = err.Error()
	} else {
		log.Printf("Response: '%v'", resp)
	}

	encoder := json.NewEncoder(c)
	if err := encoder.Encode(resp); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
