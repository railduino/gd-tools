package status

import (
	"fmt"
	"time"

	"github.com/railduino/gd-tools/config"
	"github.com/urfave/cli/v2"
)

var Command = &cli.Command{
	Name:  "status",
	Usage: "Show mTLS status and certificate info",
	Flags: []cli.Flag{
		config.FlagVerbose,
		config.FlagDry,
		config.FlagPort,
	},
	Action: Run,
}

func Run(c *cli.Context) error {
	cfg, req, err := config.ReadConfigPlus(c)
	if err != nil {
		return err
	} else if cfg != nil {
		defer cfg.Close()
	}

	state := cfg.Conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return fmt.Errorf("no peer certificate received")
	}
	cert := state.PeerCertificates[0]

	fmt.Printf("✅ Connected to %s\n", cfg.FQDN())
	fmt.Printf("📛 CN: %s\n", cert.Subject.CommonName)
	fmt.Printf("📜 SANs: %v\n", cert.DNSNames)
	fmt.Printf("🔐 Valid: %s → %s\n", cert.NotBefore.Format(time.RFC1123), cert.NotAfter.Format(time.RFC1123))

	if time.Now().After(cert.NotAfter) {
		fmt.Println("⚠️  Certificate has expired!")
	} else if time.Until(cert.NotAfter) < 7*24*time.Hour {
		fmt.Println("⚠️  Certificate expires within 7 days!")
	} else {
		fmt.Println("🟢 Certificate is valid")
	}

	req.Hello = "Status Check"

	if err := req.Send(); err != nil {
		return fmt.Errorf("agent hello failed: %w", err)
	}

	return nil
}
