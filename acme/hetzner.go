package acme

import (
	"crypto"
	"fmt"
	"os"

	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/hetzner"
	"github.com/go-acme/lego/v4/registration"
)

type HetznerUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

var HetznerNameServers = []string{
	"213.133.105.6:53",   // ns1.first-ns.de (Hetzner)
	"213.133.106.251:53", // robotns2.second-ns.de (Hetzner)
	"193.47.99.4:53",     // robotns3.second-ns.com (Hetzner)
	"8.8.8.8:53",         // Google
	"1.1.1.1:53",         // Cloudflare
}

func (u *HetznerUser) GetEmail() string {
	return u.Email
}

func (u *HetznerUser) GetRegistration() *registration.Resource {
	return u.Registration
}

func (u *HetznerUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// GetHetznerCertificate uses lego and the Hetzner DNS-01 challenge to create an ACME certificate.
// ACME is short for "Automatic Certificate Management Environment"
func GetHetznerCertificate(domains []string, email string, key crypto.PrivateKey) (*certificate.Resource, error) {
	if len(domains) == 0 {
		return nil, fmt.Errorf("missing domains in certificate request")
	}
	if os.Getenv("HETZNER_API_TOKEN") == "" {
		return nil, fmt.Errorf("missing HETZNER_API_TOKEN environment variable")
	}

	user := &HetznerUser{
		Email: email,
		key:   key,
	}

	config := lego.NewConfig(user)
	config.CADirURL = lego.LEDirectoryProduction
	config.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("unable to create lego.NewClient: %w", err)
	}

	dnsProvider, err := hetzner.NewDNSProvider()
	if err != nil {
		return nil, fmt.Errorf("failed to create hetzner.NewDNSProvider: %w", err)
	}

	challengeOpts := []dns01.ChallengeOption{
		dns01.CondOption(true, dns01.AddRecursiveNameservers(HetznerNameServers)),
	}

	if err := client.Challenge.SetDNS01Provider(dnsProvider, challengeOpts...); err != nil {
		return nil, fmt.Errorf("failed to set DNS provider: %w", err)
	}

	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, fmt.Errorf("failed to register lego client: %w", err)
	}
	user.Registration = reg

	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}

	resource, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain certificate: %w", err)
	}

	return resource, nil
}
